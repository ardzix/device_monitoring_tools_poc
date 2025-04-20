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
from django.views.generic import ListView
from django.db.models import Q
from django.db import models
from django.utils import timezone
from datetime import datetime, timedelta
from rest_framework.permissions import IsAuthenticated

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
    permission_classes = [IsAuthenticated]

class AppUsageLogViewSet(viewsets.ModelViewSet):
    queryset = AppUsageLog.objects.all().order_by('-timestamp')
    serializer_class = AppUsageLogSerializer
    permission_classes = [IsAuthenticated]

class WebsiteVisitLogViewSet(viewsets.ModelViewSet):
    queryset = WebsiteVisitLog.objects.all().order_by('-timestamp')
    serializer_class = WebsiteVisitLogSerializer
    permission_classes = [IsAuthenticated]

class FileAccessLogViewSet(viewsets.ModelViewSet):
    queryset = FileAccessLog.objects.all().order_by('-timestamp')
    serializer_class = FileAccessLogSerializer
    permission_classes = [IsAuthenticated]

class USBDeviceLogViewSet(viewsets.ModelViewSet):
    queryset = USBDeviceLog.objects.all().order_by('-timestamp')
    serializer_class = USBDeviceLogSerializer
    permission_classes = [IsAuthenticated]  # Require authentication

    def create(self, request, *args, **kwargs):
        logger.info(f"Received USB device data: {request.data}")
        return super().create(request, *args, **kwargs)

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
                logger.info(f"USB devices in data: {data.get('usb_devices', [])}")
            except json.JSONDecodeError as e:
                logger.error(f"Error parsing JSON data: {e}")
                return Response({"error": "Invalid JSON data"}, status=status.HTTP_400_BAD_REQUEST)
        else:
            data = request.data

        # Process USB device logs first
        if 'usb_devices' in data:
            logger.info(f"Processing USB device logs: {data['usb_devices']}")
            serializer = USBDeviceLogSerializer(data=data['usb_devices'], many=True)
            if serializer.is_valid():
                serializer.save()
                responses['usb_devices'] = serializer.data
                logger.info(f"Successfully saved USB device logs: {serializer.data}")
            else:
                logger.error(f"USB device validation errors: {serializer.errors}")
                return Response(serializer.errors, status=status.HTTP_400_BAD_REQUEST)

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

class LogsExplorerView(ListView):
    template_name = 'dashboard/logs_explorer.html'
    paginate_by = 50
    
    def get_queryset(self):
        # Get filter parameters from request
        log_type = self.request.GET.get('log_type', '')
        date_from = self.request.GET.get('date_from', '')
        date_to = self.request.GET.get('date_to', '')
        keyword = self.request.GET.get('keyword', '')
        flagged_only = self.request.GET.get('flagged_only', '') == 'true'
        has_screenshot = self.request.GET.get('has_screenshot', '') == 'true'
        sort_by = self.request.GET.get('sort', '-timestamp')
        
        # Convert dates to datetime objects
        try:
            date_from = datetime.strptime(date_from, '%Y-%m-%d') if date_from else timezone.now() - timedelta(days=7)
            date_to = datetime.strptime(date_to, '%Y-%m-%d') if date_to else timezone.now()
            # Make date_to inclusive of the entire day
            date_to = date_to + timedelta(days=1)
        except ValueError:
            date_from = timezone.now() - timedelta(days=7)
            date_to = timezone.now()
        
        # Base queryset for each log type
        activity_logs = ActivityLog.objects.filter(timestamp__range=(date_from, date_to))
        app_usage_logs = AppUsageLog.objects.filter(timestamp__range=(date_from, date_to))
        website_logs = WebsiteVisitLog.objects.filter(timestamp__range=(date_from, date_to))
        file_logs = FileAccessLog.objects.filter(timestamp__range=(date_from, date_to))
        usb_logs = USBDeviceLog.objects.filter(timestamp__range=(date_from, date_to))
        
        # Apply keyword filter if provided
        if keyword:
            activity_logs = activity_logs.filter(
                Q(window_title__icontains=keyword) |
                Q(keywords__icontains=keyword) |
                Q(analysis__icontains=keyword)
            )
            app_usage_logs = app_usage_logs.filter(
                Q(app_name__icontains=keyword) |
                Q(window_title__icontains=keyword)
            )
            website_logs = website_logs.filter(
                Q(url__icontains=keyword) |
                Q(title__icontains=keyword)
            )
            file_logs = file_logs.filter(
                Q(file_path__icontains=keyword) |
                Q(operation__icontains=keyword)
            )
            usb_logs = usb_logs.filter(
                Q(device_name__icontains=keyword) |
                Q(vendor_id__icontains=keyword) |
                Q(product_id__icontains=keyword)
            )
        
        # Apply additional filters
        if flagged_only:
            activity_logs = activity_logs.filter(is_flagged=True)
        if has_screenshot:
            activity_logs = activity_logs.exclude(screenshot__isnull=True).exclude(screenshot__exact='')
        
        # Combine querysets based on log type
        if log_type == 'activity':
            queryset = activity_logs
        elif log_type == 'app':
            queryset = app_usage_logs
        elif log_type == 'website':
            queryset = website_logs
        elif log_type == 'file':
            queryset = file_logs
        elif log_type == 'usb':
            queryset = usb_logs
        else:
            # Combine all logs if no specific type is selected
            queryset = activity_logs.union(
                app_usage_logs,
                website_logs,
                file_logs,
                usb_logs,
                all=True
            )
        
        # Apply sorting
        if sort_by in ['timestamp', '-timestamp']:
            queryset = queryset.order_by(sort_by)
        
        return queryset
    
    def _add_type_info(self, log, log_type):
        """Add type information to the log object"""
        if not hasattr(log, 'log_type'):
            log.log_type = log_type
        return log
    
    def get_context_data(self, **kwargs):
        context = super().get_context_data(**kwargs)
        
        # Add filter parameters to context
        context['log_type'] = self.request.GET.get('log_type', '')
        context['date_from'] = self.request.GET.get('date_from', '')
        context['date_to'] = self.request.GET.get('date_to', '')
        context['keyword'] = self.request.GET.get('keyword', '')
        context['flagged_only'] = self.request.GET.get('flagged_only', '') == 'true'
        context['has_screenshot'] = self.request.GET.get('has_screenshot', '') == 'true'
        context['sort'] = self.request.GET.get('sort', '-timestamp')
        
        # Add type information to each log
        log_type = self.request.GET.get('log_type', '')
        if log_type:
            context['object_list'] = [self._add_type_info(log, log_type) for log in context['object_list']]
        
        return context

def dashboard_view(request):
    # Get real data from models
    total_activity_logs = ActivityLog.objects.count()
    flagged_incidents = ActivityLog.objects.filter(is_flagged=True).count()
    
    # Get top keywords from activity logs
    keyword_counts = {}
    for log in ActivityLog.objects.exclude(keywords__isnull=True):
        for keyword in log.keywords:
            keyword_counts[keyword] = keyword_counts.get(keyword, 0) + 1
    top_keywords = [{'keyword': k, 'count': v} for k, v in 
                   sorted(keyword_counts.items(), key=lambda x: x[1], reverse=True)[:3]]
    
    # Get USB device count
    usb_devices = USBDeviceLog.objects.filter(action='connect').count()
    
    # Get flagged trend data (last 7 days)
    today = timezone.now().date()
    flagged_trend = {
        'labels': [(today - timedelta(days=i)).strftime('%a') for i in range(6, -1, -1)],
        'data': [
            ActivityLog.objects.filter(
                is_flagged=True,
                timestamp__date=today - timedelta(days=i)
            ).count()
            for i in range(6, -1, -1)
        ]
    }
    
    # Get risk level distribution
    risk_levels = [
        ActivityLog.objects.filter(confidence__gte=0.7).count(),  # High
        ActivityLog.objects.filter(confidence__gte=0.4, confidence__lt=0.7).count(),  # Medium
        ActivityLog.objects.filter(confidence__lt=0.4).count()  # Low
    ]
    
    # Get top websites by duration
    top_websites = WebsiteVisitLog.objects.values('url', 'title').annotate(
        total_duration=models.Sum('duration')
    ).order_by('-total_duration')[:9]
    
    # Get top apps by duration
    top_apps = AppUsageLog.objects.values('app_name', 'window_title').annotate(
        total_duration=models.Sum('duration')
    ).order_by('-total_duration')[:9]
    
    # Get top accessed files
    top_files_query = FileAccessLog.objects.values('file_path').annotate(
        count=models.Count('id')
    ).order_by('-count')[:9]
    
    top_files = [
        {
            'name': os.path.basename(item['file_path']),
            'path': item['file_path'],
            'count': item['count']
        }
        for item in top_files_query
    ]
    
    # Get recent screenshots
    recent_screenshots = ActivityLog.objects.filter(is_flagged=True).exclude(screenshot__isnull=True).exclude(screenshot='').order_by('-timestamp')[:9]
    
    context = {
        'total_activity_logs': total_activity_logs,
        'flagged_incidents': flagged_incidents,
        'top_keywords': top_keywords,
        'usb_devices': usb_devices,
        'flagged_trend': flagged_trend,
        'risk_levels': risk_levels,
        'top_websites': top_websites,
        'top_apps': top_apps,
        'top_files': top_files,
        'recent_screenshots': [
            {
                'url': f"/media/{log.screenshot}" if log.screenshot and not log.screenshot.startswith('screenshots/') else f"/media/screenshots/{os.path.basename(log.screenshot)}" if log.screenshot else None,
                'details': f"Window: {log.window_title}\nKeywords: {', '.join(log.keywords) if log.keywords else 'None'}"
            }
            for log in recent_screenshots
        ]
    }
    return render(request, 'dashboard/dashboard.html', context)

def logs_explorer_view(request):
    # Get filter parameters
    log_type = request.GET.get('log_type', '')
    date_from = request.GET.get('date_from', '')
    date_to = request.GET.get('date_to', '')
    keyword = request.GET.get('keyword', '')
    flagged_only = request.GET.get('flagged_only', '') == 'true'
    
    # Base queryset for each log type
    activity_logs = ActivityLog.objects.all()
    app_usage_logs = AppUsageLog.objects.all()
    website_logs = WebsiteVisitLog.objects.all()
    file_logs = FileAccessLog.objects.all()
    usb_logs = USBDeviceLog.objects.all()
    
    # Apply date filters if provided
    if date_from:
        try:
            date_from = datetime.strptime(date_from, '%Y-%m-%d')
            activity_logs = activity_logs.filter(timestamp__date__gte=date_from)
            app_usage_logs = app_usage_logs.filter(timestamp__date__gte=date_from)
            website_logs = website_logs.filter(timestamp__date__gte=date_from)
            file_logs = file_logs.filter(timestamp__date__gte=date_from)
            usb_logs = usb_logs.filter(timestamp__date__gte=date_from)
        except ValueError:
            pass
    
    if date_to:
        try:
            date_to = datetime.strptime(date_to, '%Y-%m-%d')
            activity_logs = activity_logs.filter(timestamp__date__lte=date_to)
            app_usage_logs = app_usage_logs.filter(timestamp__date__lte=date_to)
            website_logs = website_logs.filter(timestamp__date__lte=date_to)
            file_logs = file_logs.filter(timestamp__date__lte=date_to)
            usb_logs = usb_logs.filter(timestamp__date__lte=date_to)
        except ValueError:
            pass
    
    # Apply keyword filter if provided
    if keyword:
        activity_logs = activity_logs.filter(
            Q(window_title__icontains=keyword) |
            Q(keywords__icontains=keyword) |
            Q(analysis__icontains=keyword)
        )
        app_usage_logs = app_usage_logs.filter(
            Q(app_name__icontains=keyword) |
            Q(window_title__icontains=keyword)
        )
        website_logs = website_logs.filter(
            Q(url__icontains=keyword) |
            Q(title__icontains=keyword)
        )
        file_logs = file_logs.filter(
            Q(file_path__icontains=keyword) |
            Q(process_name__icontains=keyword)
        )
        usb_logs = usb_logs.filter(
            Q(device_name__icontains=keyword) |
            Q(vendor_id__icontains=keyword) |
            Q(product_id__icontains=keyword)
        )
    
    # Apply flagged filter if selected
    if flagged_only:
        activity_logs = activity_logs.filter(is_flagged=True)
    
    # Filter by log type if selected
    if log_type == 'activity':
        logs = activity_logs.order_by('-timestamp')
    elif log_type == 'app_usage':
        logs = app_usage_logs.order_by('-timestamp')
    elif log_type == 'website':
        logs = website_logs.order_by('-timestamp')
    elif log_type == 'file':
        logs = file_logs.order_by('-timestamp')
    elif log_type == 'usb':
        logs = usb_logs.order_by('-timestamp')
    else:
        # Combine all logs and sort by timestamp
        logs = sorted(
            list(activity_logs) +
            list(app_usage_logs) +
            list(website_logs) +
            list(file_logs) +
            list(usb_logs),
            key=lambda x: x.timestamp,
            reverse=True
        )
    
    context = {
        'logs': logs,
        'log_type': log_type,
        'date_from': date_from,
        'date_to': date_to,
        'keyword': keyword,
        'flagged_only': flagged_only
    }
    return render(request, 'dashboard/logs.html', context)
