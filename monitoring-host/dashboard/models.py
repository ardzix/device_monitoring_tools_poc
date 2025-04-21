from django.db import models
from django.utils import timezone

class BaseLog(models.Model):
    LOG_TYPES = [
        ('activity', 'Activity'),
        ('app_usage', 'App Usage'),
        ('website_visit', 'Website Visit'),
        ('file_access', 'File Access'),
        ('usb_device', 'USB Device'),
    ]

    timestamp = models.DateTimeField(auto_now_add=True)
    description = models.TextField(blank=True, null=True)
    log_type = models.CharField(max_length=20, choices=LOG_TYPES)
    device_identifier = models.CharField(max_length=255, help_text="Unique identifier for the device")

    class Meta:
        ordering = ['-timestamp']
        verbose_name = 'Base Log'
        verbose_name_plural = 'Base Logs'

    def __str__(self):
        return f"{self.timestamp} - {self.device_identifier} ({self.log_type})"

class ActivityLog(BaseLog):
    window_title = models.CharField(max_length=255)
    clipboard = models.TextField(blank=True)
    screenshot = models.ImageField(upload_to='screenshots/', blank=True, null=True)
    is_flagged = models.BooleanField(default=False)
    confidence = models.FloatField(default=0.0)
    analysis = models.TextField(blank=True)
    keywords = models.JSONField(default=list, blank=True, null=True)

    def save(self, *args, **kwargs):
        self.log_type = 'activity'
        # Generate a descriptive summary
        flag_status = "🚩 Flagged" if self.is_flagged else "✓ Normal"
        keywords_info = ""
        if self.keywords and len(self.keywords) > 0:
            keywords_str = ", ".join(self.keywords[:3])  # Show first 3 keywords
            if len(self.keywords) > 3:
                keywords_str += f" (+{len(self.keywords) - 3} more)"
            keywords_info = f" [Keywords: {keywords_str}]"
        
        if self.screenshot:
            self.description = f"{flag_status} activity in '{self.window_title}' (with screenshot){keywords_info}"
        else:
            self.description = f"{flag_status} activity in '{self.window_title}'{keywords_info}"
        super().save(*args, **kwargs)

    def __str__(self):
        flag_status = "FLAGGED" if self.is_flagged else "Normal"
        return f"{self.timestamp} - {self.window_title} ({flag_status})"

class AppUsageLog(BaseLog):
    app_name = models.CharField(max_length=255)
    window_title = models.CharField(max_length=255)
    duration = models.IntegerField(default=0)  # Duration in seconds
    is_active = models.BooleanField(default=False)

    def save(self, *args, **kwargs):
        self.log_type = 'app_usage'
        # Format duration for description
        minutes = self.duration // 60
        seconds = self.duration % 60
        duration_str = f"{minutes}m {seconds}s" if minutes > 0 else f"{seconds}s"
        status = "Active" if self.is_active else "Inactive"
        self.description = f"{status}: {self.app_name} - '{self.window_title}' ({duration_str})"
        super().save(*args, **kwargs)

    def __str__(self):
        status = "Active" if self.is_active else "Inactive"
        return f"{self.timestamp} - {self.app_name} ({status}, {self.duration}s)"

class WebsiteVisitLog(BaseLog):
    url = models.URLField()
    title = models.CharField(max_length=255)
    duration = models.IntegerField(default=0)  # Duration in seconds

    def save(self, *args, **kwargs):
        self.log_type = 'website_visit'
        # Format duration for description
        minutes = self.duration // 60
        seconds = self.duration % 60
        duration_str = f"{minutes}m {seconds}s" if minutes > 0 else f"{seconds}s"
        self.description = f"Visited '{self.title}' ({duration_str})"
        super().save(*args, **kwargs)

    def __str__(self):
        return f"{self.timestamp} - {self.title} ({self.duration}s)"

class FileAccessLog(BaseLog):
    file_path = models.CharField(max_length=512)
    operation = models.CharField(max_length=50)  # create, modify, delete, read
    process_name = models.CharField(max_length=255)

    def save(self, *args, **kwargs):
        self.log_type = 'file_access'
        operation_icons = {
            'create': '📝',
            'modify': '✏️',
            'delete': '🗑️',
            'read': '👁️'
        }
        icon = operation_icons.get(self.operation.lower(), '❓')
        self.description = f"{icon} {self.operation.title()} '{self.file_path}' by {self.process_name}"
        super().save(*args, **kwargs)

    def __str__(self):
        return f"{self.timestamp} - {self.operation} {self.file_path}"

class USBDeviceLog(BaseLog):
    device_name = models.CharField(max_length=255)
    vendor_id = models.CharField(max_length=10)
    product_id = models.CharField(max_length=10)
    serial_number = models.CharField(max_length=255, blank=True)
    action = models.CharField(max_length=50)  # connected, disconnected

    def save(self, *args, **kwargs):
        self.log_type = 'usb_device'
        icon = '🔌' if self.action.lower() == 'connected' else '🔌❌'
        self.description = f"{icon} {self.action.title()}: {self.device_name} ({self.vendor_id}:{self.product_id})"
        if self.serial_number:
            self.description += f" S/N: {self.serial_number}"
        super().save(*args, **kwargs)

    def __str__(self):
        return f"{self.timestamp} - {self.device_name} ({self.action})"
