from django.shortcuts import render
from rest_framework import viewsets, filters, permissions, status
from django_filters.rest_framework import DjangoFilterBackend
from rest_framework.decorators import action
from rest_framework.response import Response
from .models import ActivityLog, AppUsageLog, WebsiteVisitLog, FileAccessLog, USBDeviceLog
from .serializers import (
    ActivityLogSerializer,
    AppUsageLogSerializer,
    WebsiteVisitLogSerializer,
    FileAccessLogSerializer,
    USBDeviceLogSerializer
)
import json
from django.core.files.storage import default_storage
from django.core.files.base import ContentFile
import logging
import os

logger = logging.getLogger(__name__)

# Create your views here.

class ActivityLogViewSet(viewsets.ModelViewSet):
    queryset = ActivityLog.objects.all().order_by('-timestamp')
    serializer_class = ActivityLogSerializer
    filter_backends = [DjangoFilterBackend, filters.SearchFilter, filters.OrderingFilter]
    filterset_fields = ['is_flagged']
    search_fields = ['window_title', 'clipboard_content', 'analysis']
    ordering_fields = ['timestamp', 'confidence']
    ordering = ['-timestamp']
    permission_classes = []  # Allow any requests

class AppUsageLogViewSet(viewsets.ModelViewSet):
    queryset = AppUsageLog.objects.all().order_by('-timestamp')
    serializer_class = AppUsageLogSerializer
    permission_classes = []  # Allow any requests

class WebsiteVisitLogViewSet(viewsets.ModelViewSet):
    queryset = WebsiteVisitLog.objects.all().order_by('-timestamp')
    serializer_class = WebsiteVisitLogSerializer
    permission_classes = []  # Allow any requests

class FileAccessLogViewSet(viewsets.ModelViewSet):
    queryset = FileAccessLog.objects.all().order_by('-timestamp')
    serializer_class = FileAccessLogSerializer
    permission_classes = []  # Allow any requests

class USBDeviceLogViewSet(viewsets.ModelViewSet):
    queryset = USBDeviceLog.objects.all().order_by('-timestamp')
    serializer_class = USBDeviceLogSerializer
    permission_classes = []  # Allow any requests

class BulkMonitoringViewSet(viewsets.ViewSet):
    def create(self, request):
        logger.info("Received bulk data request")
        logger.info(f"Files in request: {request.FILES.keys()}")
        logger.info(f"Data in request: {request.data.keys()}")

        responses = {
            'app_usage': [],
            'website_visits': [],
            'file_access': [],
            'usb_devices': [],
            'activity_logs': []
        }

        # Handle multipart form data
        if 'data' in request.data:
            try:
                data = json.loads(request.data['data'])
                logger.info("Successfully parsed JSON data")
            except json.JSONDecodeError as e:
                logger.error(f"Error parsing JSON data: {e}")
                return Response({"error": "Invalid JSON data"}, status=status.HTTP_400_BAD_REQUEST)
        else:
            data = request.data

        # Process app usage logs
        if 'app_usage' in data:
            serializer = AppUsageLogSerializer(data=data['app_usage'], many=True)
            if serializer.is_valid():
                serializer.save()
                responses['app_usage'] = serializer.data
            else:
                return Response(serializer.errors, status=status.HTTP_400_BAD_REQUEST)

        # Process website visit logs
        if 'website_visits' in data:
            serializer = WebsiteVisitLogSerializer(data=data['website_visits'], many=True)
            if serializer.is_valid():
                serializer.save()
                responses['website_visits'] = serializer.data
            else:
                return Response(serializer.errors, status=status.HTTP_400_BAD_REQUEST)

        # Process file access logs
        if 'file_access' in data:
            serializer = FileAccessLogSerializer(data=data['file_access'], many=True)
            if serializer.is_valid():
                serializer.save()
                responses['file_access'] = serializer.data
            else:
                return Response(serializer.errors, status=status.HTTP_400_BAD_REQUEST)

        # Process USB device logs
        if 'usb_devices' in data:
            serializer = USBDeviceLogSerializer(data=data['usb_devices'], many=True)
            if serializer.is_valid():
                serializer.save()
                responses['usb_devices'] = serializer.data
            else:
                return Response(serializer.errors, status=status.HTTP_400_BAD_REQUEST)

        # Process activity logs
        if 'activity_logs' in data:
            activity_logs = []
            
            for log_data in data['activity_logs']:
                log_copy = log_data.copy()  # Create a copy to avoid modifying the original
                
                # Handle screenshot file if present
                if 'screenshot' in log_copy and log_copy['screenshot']:
                    original_screenshot_path = log_copy['screenshot']
                    if 'screenshot' in request.FILES:
                        file = request.FILES['screenshot']
                        filename = os.path.basename(original_screenshot_path)
                        try:
                            saved_path = default_storage.save(f'screenshots/{filename}', ContentFile(file.read()))
                            log_copy['screenshot'] = saved_path  # Update with the saved path
                            logger.info(f"Saved screenshot to {saved_path}")
                        except Exception as e:
                            logger.error(f"Error saving screenshot: {e}")
                            log_copy['screenshot'] = None
                    else:
                        logger.warning(f"Screenshot file not found in request.FILES. Available fields: {list(request.FILES.keys())}")
                        log_copy['screenshot'] = None

                activity_logs.append(log_copy)

            logger.info(f"Processing activity logs: {activity_logs}")
            serializer = ActivityLogSerializer(data=activity_logs, many=True)
            if serializer.is_valid():
                serializer.save()
                responses['activity_logs'] = serializer.data
            else:
                logger.error(f"Activity log validation errors: {serializer.errors}")
                return Response(serializer.errors, status=status.HTTP_400_BAD_REQUEST)

        return Response(responses, status=status.HTTP_201_CREATED)
