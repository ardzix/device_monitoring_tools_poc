from rest_framework import serializers
from .models import ActivityLog, AppUsageLog, WebsiteVisitLog, FileAccessLog, USBDeviceLog
import json
import logging

logger = logging.getLogger(__name__)

class ActivityLogSerializer(serializers.ModelSerializer):
    screenshot = serializers.CharField(allow_null=True, required=False)
    keywords = serializers.ListField(child=serializers.CharField(), allow_empty=True, required=False)

    class Meta:
        model = ActivityLog
        fields = ['timestamp', 'description', 'device_identifier', 'window_title', 'clipboard', 
                 'screenshot', 'analysis', 'is_flagged', 'keywords', 'confidence']
        read_only_fields = ('id', 'timestamp')

    def validate_screenshot(self, value):
        if value is None or isinstance(value, str):
            return value
        return str(value)

    def validate_keywords(self, value):
        if isinstance(value, str):
            try:
                return json.loads(value)
            except json.JSONDecodeError:
                return []
        return value

    def create(self, validated_data):
        try:
            # Handle the screenshot file upload
            screenshot = validated_data.pop('screenshot', None)
            
            # Create the instance
            instance = super().create(validated_data)
            
            # Save the screenshot if it exists
            if screenshot:
                instance.screenshot = screenshot
                instance.save()
                
            return instance
        except Exception as e:
            print(f"Error in serializer create: {str(e)}")
            raise

class AppUsageLogSerializer(serializers.ModelSerializer):
    class Meta:
        model = AppUsageLog
        fields = ['timestamp', 'description', 'device_identifier', 'app_name', 'window_title', 
                 'duration', 'is_active']

class WebsiteVisitLogSerializer(serializers.ModelSerializer):
    class Meta:
        model = WebsiteVisitLog
        fields = ['timestamp', 'description', 'device_identifier', 'url', 'title', 'duration']

class FileAccessLogSerializer(serializers.ModelSerializer):
    class Meta:
        model = FileAccessLog
        fields = ['timestamp', 'description', 'device_identifier', 'file_path', 'operation', 
                 'process_name']

class USBDeviceLogSerializer(serializers.ModelSerializer):
    class Meta:
        model = USBDeviceLog
        fields = ['timestamp', 'description', 'device_identifier', 'device_name', 'vendor_id', 
                 'product_id', 'serial_number', 'action']

    def to_internal_value(self, data):
        # Map 'name' to 'device_name' if present
        if 'name' in data:
            data['device_name'] = data.pop('name')
        return super().to_internal_value(data)

class BulkMonitoringSerializer(serializers.Serializer):
    app_usage = AppUsageLogSerializer(many=True, required=False)
    website_visits = WebsiteVisitLogSerializer(many=True, required=False)
    file_access = FileAccessLogSerializer(many=True, required=False)
    usb_devices = USBDeviceLogSerializer(many=True, required=False)
    activity_logs = ActivityLogSerializer(many=True, required=False)

    def validate(self, data):
        logger.info(f"Validating bulk data: {data}")
        return data

    def validate_app_usage(self, value):
        logger.info(f"Validating app usage data: {value}")
        return value

    def validate_website_visits(self, value):
        logger.info(f"Validating website visits data: {value}")
        return value

    def validate_file_access(self, value):
        logger.info(f"Validating file access data: {value}")
        return value

    def validate_usb_devices(self, value):
        logger.info(f"Validating USB devices data: {value}")
        return value

    def validate_activity_logs(self, value):
        logger.info(f"Validating activity logs data: {value}")
        for log in value:
            if 'screenshot' in log:
                # Remove the screenshot field from the data as it will be handled separately
                del log['screenshot']
        return value

    def to_internal_value(self, data):
        logger.info(f"Converting to internal value: {data}")
        ret = super().to_internal_value(data)
        logger.info(f"Converted to internal value: {ret}")
        return ret 