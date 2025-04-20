from django.contrib import admin
from .models import ActivityLog, AppUsageLog, WebsiteVisitLog, FileAccessLog, USBDeviceLog

@admin.register(ActivityLog)
class ActivityLogAdmin(admin.ModelAdmin):
    list_display = ('timestamp', 'window_title', 'is_flagged', 'confidence')
    list_filter = ('is_flagged', 'timestamp')
    search_fields = ('window_title', 'clipboard', 'analysis')
    readonly_fields = ('timestamp',)
    ordering = ('-timestamp',)

@admin.register(AppUsageLog)
class AppUsageLogAdmin(admin.ModelAdmin):
    list_display = ('timestamp', 'app_name', 'window_title', 'duration', 'is_active')
    list_filter = ('is_active', 'app_name', 'timestamp')
    search_fields = ('app_name', 'window_title')
    readonly_fields = ('timestamp',)
    ordering = ('-timestamp',)

@admin.register(WebsiteVisitLog)
class WebsiteVisitLogAdmin(admin.ModelAdmin):
    list_display = ('timestamp', 'url', 'title', 'duration')
    list_filter = ('timestamp',)
    search_fields = ('url', 'title')
    readonly_fields = ('timestamp',)
    ordering = ('-timestamp',)

@admin.register(FileAccessLog)
class FileAccessLogAdmin(admin.ModelAdmin):
    list_display = ('timestamp', 'file_path', 'operation', 'process_name')
    list_filter = ('operation', 'process_name', 'timestamp')
    search_fields = ('file_path', 'process_name')
    readonly_fields = ('timestamp',)
    ordering = ('-timestamp',)

@admin.register(USBDeviceLog)
class USBDeviceLogAdmin(admin.ModelAdmin):
    list_display = ('timestamp', 'device_name', 'vendor_id', 'product_id', 'action')
    list_filter = ('action', 'vendor_id', 'timestamp')
    search_fields = ('device_name', 'serial_number')
    readonly_fields = ('timestamp',)
    ordering = ('-timestamp',)
