from rest_framework import serializers
from .models import ActivityLog, AppUsageLog, WebsiteVisitLog, FileAccessLog, USBDeviceLog
import json

class ActivityLogSerializer(serializers.ModelSerializer):
    screenshot = serializers.CharField(allow_null=True, required=False)
    keywords = serializers.ListField(child=serializers.CharField(), allow_empty=True, required=False)

    class Meta:
        model = ActivityLog
        fields = '__all__'
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
        fields = '__all__'

class WebsiteVisitLogSerializer(serializers.ModelSerializer):
    class Meta:
        model = WebsiteVisitLog
        fields = '__all__'

class FileAccessLogSerializer(serializers.ModelSerializer):
    class Meta:
        model = FileAccessLog
        fields = '__all__'

class USBDeviceLogSerializer(serializers.ModelSerializer):
    class Meta:
        model = USBDeviceLog
        fields = '__all__'

    def to_internal_value(self, data):
        # Map 'name' to 'device_name' if present
        if 'name' in data:
            data['device_name'] = data.pop('name')
        return super().to_internal_value(data) 