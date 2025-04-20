from django.db import models
from django.utils import timezone

# Create your models here.

class ActivityLog(models.Model):
    timestamp = models.DateTimeField(auto_now_add=True)
    window_title = models.CharField(max_length=255, blank=True, null=True)
    clipboard = models.TextField(blank=True, null=True)
    screenshot = models.CharField(max_length=255, blank=True, null=True)  # Store path to screenshot
    analysis = models.TextField(blank=True, null=True)
    is_flagged = models.BooleanField(default=False)
    keywords = models.JSONField(default=list, blank=True, null=True)
    confidence = models.FloatField(default=0.0)

    def __str__(self):
        return f"{self.timestamp} - {self.window_title}"

class AppUsageLog(models.Model):
    timestamp = models.DateTimeField(auto_now_add=True)
    app_name = models.CharField(max_length=255)
    window_title = models.CharField(max_length=255)
    duration = models.IntegerField()
    is_active = models.BooleanField(default=True)

    def __str__(self):
        return f"{self.timestamp} - {self.app_name}"

class WebsiteVisitLog(models.Model):
    timestamp = models.DateTimeField(auto_now_add=True)
    url = models.URLField()
    title = models.CharField(max_length=255)
    duration = models.IntegerField()

    def __str__(self):
        return f"{self.timestamp} - {self.url}"

class FileAccessLog(models.Model):
    timestamp = models.DateTimeField(auto_now_add=True)
    file_path = models.CharField(max_length=255)
    operation = models.CharField(max_length=50)
    process_name = models.CharField(max_length=255)

    def __str__(self):
        return f"{self.timestamp} - {self.file_path}"

class USBDeviceLog(models.Model):
    timestamp = models.DateTimeField(auto_now_add=True)
    device_name = models.CharField(max_length=255)
    vendor_id = models.CharField(max_length=10)
    product_id = models.CharField(max_length=10)
    serial_number = models.CharField(max_length=50, null=True, blank=True)
    action = models.CharField(max_length=50)

    def __str__(self):
        return f"{self.timestamp} - {self.device_name}"
