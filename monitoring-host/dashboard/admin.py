from django.contrib import admin
from django.utils.html import format_html
from django.urls import reverse
from .models import BaseLog, ActivityLog, AppUsageLog, WebsiteVisitLog, FileAccessLog, USBDeviceLog

@admin.register(BaseLog)
class BaseLogAdmin(admin.ModelAdmin):
    list_display = ('timestamp', 'device_identifier', 'log_type', 'description', 'get_details_link')
    list_filter = ('log_type', 'device_identifier', 'timestamp')
    search_fields = ('device_identifier', 'description')
    readonly_fields = ('timestamp', 'log_type')
    ordering = ('-timestamp',)
    list_per_page = 50
    date_hierarchy = 'timestamp'

    def get_details_link(self, obj):
        # Map of log types to their model names for URL construction
        model_map = {
            'activity': ('activitylog', ActivityLog),
            'app_usage': ('appusagelog', AppUsageLog),
            'website_visit': ('websitevisitlog', WebsiteVisitLog),
            'file_access': ('fileaccesslog', FileAccessLog),
            'usb_device': ('usbdevicelog', USBDeviceLog),
        }
        
        if obj.log_type in model_map:
            # Get the model name and class
            model_name, model_class = model_map[obj.log_type]
            try:
                specific_obj = model_class.objects.get(baselog_ptr_id=obj.id)
                url = reverse(f'admin:dashboard_{model_name}_change', args=[specific_obj.id])
                return format_html('<a href="{}" target="_blank">View Details</a>', url)
            except model_class.DoesNotExist:
                return ''
        return ''
    get_details_link.short_description = 'Details'

@admin.register(ActivityLog)
class ActivityLogAdmin(admin.ModelAdmin):
    list_display = ('timestamp', 'device_identifier', 'window_title', 'has_screenshot_link', 'is_flagged', 'confidence', 'colored_analysis')
    list_filter = ('is_flagged', 'device_identifier', 'timestamp')
    search_fields = ('window_title', 'clipboard', 'analysis', 'device_identifier')
    readonly_fields = ('timestamp', 'log_type', 'colored_analysis', 'has_screenshot_link')
    ordering = ('-timestamp',)
    list_per_page = 50
    date_hierarchy = 'timestamp'

    def has_screenshot(self, obj):
        return bool(obj.screenshot)
    has_screenshot.boolean = True
    has_screenshot.short_description = 'Has Screenshot'

    def has_screenshot_link(self, obj):
        if obj.screenshot:
            return format_html('<a href="/media/{}" target="_blank">View Screenshot</a>', obj.screenshot)
        return '—'
    has_screenshot_link.short_description = 'Screenshot'

    def colored_analysis(self, obj):
        if obj.is_flagged:
            return format_html('<span style="color: red;">{}</span>', obj.analysis)
        return obj.analysis
    colored_analysis.short_description = 'Analysis'

@admin.register(AppUsageLog)
class AppUsageLogAdmin(admin.ModelAdmin):
    list_display = ('timestamp', 'device_identifier', 'app_name', 'window_title', 'formatted_duration', 'active_status')
    list_filter = ('is_active', 'app_name', 'device_identifier', 'timestamp')
    search_fields = ('app_name', 'window_title', 'device_identifier')
    readonly_fields = ('timestamp', 'log_type', 'formatted_duration')
    ordering = ('-timestamp',)
    list_per_page = 50
    date_hierarchy = 'timestamp'

    def formatted_duration(self, obj):
        minutes = obj.duration // 60
        seconds = obj.duration % 60
        if minutes > 0:
            return f"{minutes}m {seconds}s"
        return f"{seconds}s"
    formatted_duration.short_description = 'Duration'

    def active_status(self, obj):
        color = 'green' if obj.is_active else 'red'
        status = 'Active' if obj.is_active else 'Inactive'
        return format_html('<span style="color: {};">{}</span>', color, status)
    active_status.short_description = 'Status'

@admin.register(WebsiteVisitLog)
class WebsiteVisitLogAdmin(admin.ModelAdmin):
    list_display = ('timestamp', 'device_identifier', 'title', 'formatted_url', 'formatted_duration')
    list_filter = ('device_identifier', 'timestamp')
    search_fields = ('url', 'title', 'device_identifier')
    readonly_fields = ('timestamp', 'log_type', 'formatted_duration')
    ordering = ('-timestamp',)
    list_per_page = 50
    date_hierarchy = 'timestamp'

    def formatted_url(self, obj):
        return format_html('<a href="{}" target="_blank">{}</a>', obj.url, obj.url[:50] + '...' if len(obj.url) > 50 else obj.url)
    formatted_url.short_description = 'URL'

    def formatted_duration(self, obj):
        minutes = obj.duration // 60
        seconds = obj.duration % 60
        if minutes > 0:
            return f"{minutes}m {seconds}s"
        return f"{seconds}s"
    formatted_duration.short_description = 'Duration'

@admin.register(FileAccessLog)
class FileAccessLogAdmin(admin.ModelAdmin):
    list_display = ('timestamp', 'device_identifier', 'file_path', 'colored_operation', 'process_name')
    list_filter = ('operation', 'process_name', 'device_identifier', 'timestamp')
    search_fields = ('file_path', 'process_name', 'device_identifier')
    readonly_fields = ('timestamp', 'log_type')
    ordering = ('-timestamp',)
    list_per_page = 50
    date_hierarchy = 'timestamp'

    def colored_operation(self, obj):
        colors = {
            'create': 'green',
            'modify': 'orange',
            'delete': 'red',
            'read': 'blue'
        }
        color = colors.get(obj.operation.lower(), 'black')
        return format_html('<span style="color: {};">{}</span>', color, obj.operation)
    colored_operation.short_description = 'Operation'

@admin.register(USBDeviceLog)
class USBDeviceLogAdmin(admin.ModelAdmin):
    list_display = ('timestamp', 'device_identifier', 'device_name', 'vendor_id', 'product_id', 'serial_number', 'colored_action')
    list_filter = ('action', 'vendor_id', 'device_identifier', 'timestamp')
    search_fields = ('device_name', 'serial_number', 'device_identifier', 'vendor_id', 'product_id')
    readonly_fields = ('timestamp', 'log_type')
    ordering = ('-timestamp',)
    list_per_page = 50
    date_hierarchy = 'timestamp'

    def colored_action(self, obj):
        colors = {
            'connected': 'green',
            'disconnected': 'red',
        }
        color = colors.get(obj.action.lower(), 'black')
        return format_html('<span style="color: {};">{}</span>', color, obj.action)
    colored_action.short_description = 'Action'
