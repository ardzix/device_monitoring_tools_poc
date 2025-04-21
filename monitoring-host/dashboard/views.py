from django.shortcuts import render
from rest_framework import viewsets, filters, permissions, status
from django_filters.rest_framework import DjangoFilterBackend
from rest_framework.decorators import action
from rest_framework.response import Response
from .models import ActivityLog, AppUsageLog, WebsiteVisitLog, FileAccessLog, USBDeviceLog, BaseLog
from .serializers import (
    ActivityLogSerializer,
    AppUsageLogSerializer,
    WebsiteVisitLogSerializer,
    FileAccessLogSerializer,
    USBDeviceLogSerializer,
    BulkMonitoringSerializer
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
from django.conf import settings
from django.core.paginator import Paginator, PageNotAnInteger, EmptyPage

logger = logging.getLogger(__name__)

# Create your views here.

class ActivityLogViewSet(viewsets.ModelViewSet):
    queryset = ActivityLog.objects.all()
    serializer_class = ActivityLogSerializer
    permission_classes = [permissions.AllowAny]

class AppUsageLogViewSet(viewsets.ModelViewSet):
    queryset = AppUsageLog.objects.all()
    serializer_class = AppUsageLogSerializer
    permission_classes = [permissions.AllowAny]

class WebsiteVisitLogViewSet(viewsets.ModelViewSet):
    queryset = WebsiteVisitLog.objects.all()
    serializer_class = WebsiteVisitLogSerializer
    permission_classes = [permissions.AllowAny]

class FileAccessLogViewSet(viewsets.ModelViewSet):
    queryset = FileAccessLog.objects.all()
    serializer_class = FileAccessLogSerializer
    permission_classes = [permissions.AllowAny]

class USBDeviceLogViewSet(viewsets.ModelViewSet):
    queryset = USBDeviceLog.objects.all()
    serializer_class = USBDeviceLogSerializer
    permission_classes = [permissions.AllowAny]

class BulkMonitoringViewSet(viewsets.ModelViewSet):
    queryset = ActivityLog.objects.all()
    serializer_class = BulkMonitoringSerializer
    permission_classes = [permissions.AllowAny]

    def create(self, request, *args, **kwargs):
        logger.info(f"Received bulk data: {request.data}")
        logger.info(f"Received files: {request.FILES}")
        
        # Parse the JSON data from the request
        try:
            data = json.loads(request.data.get('data', '{}'))
            logger.info(f"Parsed JSON data: {data}")
        except json.JSONDecodeError as e:
            logger.error(f"Error parsing JSON data: {e}")
            return Response({"error": "Invalid JSON data"}, status=400)

        # Create serializer with parsed data
        serializer = self.get_serializer(data=data)
        if not serializer.is_valid():
            logger.error(f"Serializer errors: {serializer.errors}")
            return Response(serializer.errors, status=400)

        logger.info(f"Validated data: {serializer.validated_data}")
        
        # Process each type of log
        try:
            # Process app usage logs
            if 'app_usage' in serializer.validated_data:
                for log_data in serializer.validated_data['app_usage']:
                    logger.info(f"Creating app usage log: {log_data}")
                    AppUsageLog.objects.create(**log_data)

            # Process website visit logs
            if 'website_visits' in serializer.validated_data:
                for log_data in serializer.validated_data['website_visits']:
                    logger.info(f"Creating website visit log: {log_data}")
                    WebsiteVisitLog.objects.create(**log_data)

            # Process file access logs
            if 'file_access' in serializer.validated_data:
                for log_data in serializer.validated_data['file_access']:
                    logger.info(f"Creating file access log: {log_data}")
                    FileAccessLog.objects.create(**log_data)

            # Process USB device logs
            if 'usb_devices' in serializer.validated_data:
                for log_data in serializer.validated_data['usb_devices']:
                    logger.info(f"Creating USB device log: {log_data}")
                    USBDeviceLog.objects.create(**log_data)

            # Process activity logs
            if 'activity_logs' in serializer.validated_data:
                for log_data in serializer.validated_data['activity_logs']:
                    logger.info(f"Creating activity log: {log_data}")
                    # Handle screenshot file if present
                    if 'screenshot' in log_data:
                        screenshot_path = log_data['screenshot']
                        filename = os.path.basename(screenshot_path)
                        
                        # Check if the file was uploaded in the request
                        if filename in request.FILES:
                            # Save the file to MEDIA_ROOT/screenshots
                            file_content = request.FILES[filename]
                            save_path = os.path.join('screenshots', filename)
                            path = default_storage.save(save_path, file_content)
                            log_data['screenshot'] = path
                            logger.info(f"Saved screenshot to {path}")
                        else:
                            logger.warning(f"Screenshot file {filename} not found in request.FILES")
                            log_data['screenshot'] = ''  # Clear the path if file wasn't uploaded
                    
                    ActivityLog.objects.create(**log_data)

            return Response(status=201)
        except Exception as e:
            logger.error(f"Error creating logs: {e}")
            return Response({"error": str(e)}, status=500)

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
    # Get total counts for each log type
    total_activity_logs = ActivityLog.objects.count()
    total_app_usage = AppUsageLog.objects.count()
    total_website_visits = WebsiteVisitLog.objects.count()
    total_file_access = FileAccessLog.objects.count()
    total_usb_events = USBDeviceLog.objects.count()

    # Get device statistics
    devices = BaseLog.objects.values('device_identifier').distinct()
    total_devices = devices.count()
    
    # Get activity statistics per device
    device_stats = []
    for device in devices:
        device_id = device['device_identifier']
        stats = {
            'device_identifier': device_id,
            'activity_count': ActivityLog.objects.filter(device_identifier=device_id).count(),
            'flagged_count': ActivityLog.objects.filter(device_identifier=device_id, is_flagged=True).count(),
            'app_usage_count': AppUsageLog.objects.filter(device_identifier=device_id).count(),
            'website_visits': WebsiteVisitLog.objects.filter(device_identifier=device_id).count(),
            'file_operations': FileAccessLog.objects.filter(device_identifier=device_id).count(),
            'usb_events': USBDeviceLog.objects.filter(device_identifier=device_id).count(),
            'last_seen': BaseLog.objects.filter(device_identifier=device_id).order_by('-timestamp').first().timestamp,
        }
        device_stats.append(stats)

    # Get recent flagged activities
    recent_flagged = ActivityLog.objects.filter(
        is_flagged=True
    ).order_by('-timestamp')[:10]

    # Get top keywords across all devices
    keyword_stats = {}
    for log in ActivityLog.objects.exclude(keywords__isnull=True):
        if log.keywords:
            for keyword in log.keywords:
                if keyword in keyword_stats:
                    keyword_stats[keyword] += 1
                else:
                    keyword_stats[keyword] = 1
    
    top_keywords = sorted(
        [{'keyword': k, 'count': v} for k, v in keyword_stats.items()],
        key=lambda x: x['count'],
        reverse=True
    )[:10]

    # Get recent screenshots with context
    recent_screenshots = ActivityLog.objects.filter(
        screenshot__isnull=False,
        is_flagged=True
    ).exclude(
        screenshot=''
    ).order_by('-timestamp')[:6]

    # Get active hours distribution (last 24 hours)
    now = timezone.now()
    last_24h = now - timedelta(hours=24)
    
    # Create a list of all hours
    hourly_activity = []
    for hour in range(24):
        hour_start = last_24h.replace(hour=hour, minute=0, second=0, microsecond=0)
        hour_end = hour_start + timedelta(hours=1)
        count = BaseLog.objects.filter(
            timestamp__gte=hour_start,
            timestamp__lt=hour_end
        ).count()
        hourly_activity.append({'hour': hour, 'count': count})

    # Get file operations summary
    file_operations = FileAccessLog.objects.values('operation').annotate(
        count=models.Count('id')
    ).order_by('-count')

    # Get USB device summary
    usb_summary = USBDeviceLog.objects.values('action').annotate(
        count=models.Count('id')
    ).order_by('-count')

    # Get top accessed websites
    top_websites = WebsiteVisitLog.objects.values('url', 'title').annotate(
        visit_count=models.Count('id'),
        total_duration=models.Sum('duration')
    ).order_by('-visit_count')[:5]

    # Get top used applications
    top_apps = AppUsageLog.objects.values('app_name').annotate(
        usage_count=models.Count('id'),
        total_duration=models.Sum('duration'),
        active_time=models.Sum(models.Case(
            models.When(is_active=True, then=models.F('duration')),
            default=0,
            output_field=models.IntegerField(),
        ))
    ).order_by('-total_duration')[:5]

    context = {
        # Overview statistics
        'total_activity_logs': total_activity_logs,
        'total_app_usage': total_app_usage,
        'total_website_visits': total_website_visits,
        'total_file_access': total_file_access,
        'total_usb_events': total_usb_events,
        'total_devices': total_devices,
        
        # Device-specific statistics
        'device_stats': device_stats,
        
        # Recent flagged activities
        'recent_flagged': recent_flagged,
        
        # Keyword analysis
        'top_keywords': top_keywords,
        
        # Screenshots with context
        'recent_screenshots': [
            {
                'timestamp': log.timestamp,
                'window_title': log.window_title,
                'url': log.screenshot.url if log.screenshot else None,
                'is_flagged': log.is_flagged,
                'confidence': log.confidence,
                'analysis': log.analysis,
                'keywords': log.keywords,
                'device_identifier': log.device_identifier,
            } for log in recent_screenshots
        ],
        
        # Activity distribution
        'hourly_activity': hourly_activity,
        
        # File operations summary
        'file_operations': file_operations,
        
        # USB activity summary
        'usb_summary': usb_summary,
        
        # Top websites
        'top_websites': [
            {
                'url': site['url'],
                'title': site['title'],
                'visit_count': site['visit_count'],
                'total_duration': timedelta(seconds=site['total_duration'])
            } for site in top_websites
        ],
        
        # Top applications
        'top_apps': [
            {
                'name': app['app_name'],
                'usage_count': app['usage_count'],
                'total_duration': timedelta(seconds=app['total_duration']),
                'active_time': timedelta(seconds=app['active_time']),
                'idle_time': timedelta(seconds=app['total_duration'] - app['active_time'])
            } for app in top_apps
        ],
    }
    
    return render(request, 'dashboard/dashboard.html', context)

def logs_explorer_view(request):
    # Get filter parameters
    log_type = request.GET.get('log_type', '')
    date_from = request.GET.get('date_from', '')
    date_to = request.GET.get('date_to', '')
    keyword = request.GET.get('keyword', '')
    flagged_only = request.GET.get('flagged_only', '') == 'true'
    has_screenshot = request.GET.get('has_screenshot', '') == 'true'
    sort_by = request.GET.get('sort', '-timestamp')
    order = request.GET.get('order', 'desc')
    page = request.GET.get('page', 1)
    per_page = 50

    # Base queryset with select_related for better performance
    queryset = BaseLog.objects.select_related(
        'activitylog',
        'appusagelog',
        'websitevisitlog',
        'fileaccesslog',
        'usbdevicelog'
    )

    # Apply date filters if provided
    if date_from:
        try:
            date_from = datetime.strptime(date_from, '%Y-%m-%d')
            queryset = queryset.filter(timestamp__date__gte=date_from)
        except ValueError:
            pass
    
    if date_to:
        try:
            date_to = datetime.strptime(date_to, '%Y-%m-%d')
            queryset = queryset.filter(timestamp__date__lte=date_to)
        except ValueError:
            pass

    # Apply log type filter
    if log_type:
        queryset = queryset.filter(log_type=log_type)

    # Apply keyword filter
    if keyword:
        queryset = queryset.filter(
            Q(description__icontains=keyword) |
            Q(device_identifier__icontains=keyword)
        )

    # Apply flagged filter (only for activity logs)
    if flagged_only:
        queryset = queryset.filter(
            Q(log_type='activity', activitylog__is_flagged=True)
        )

    # Apply screenshot filter (only for activity logs)
    if has_screenshot:
        queryset = queryset.filter(
            Q(log_type='activity', activitylog__screenshot__isnull=False)
        ).exclude(
            Q(log_type='activity', activitylog__screenshot='')
        )

    # Apply sorting
    if sort_by.lstrip('-') in ['timestamp', 'device_identifier']:
        if order == 'asc' and sort_by.startswith('-'):
            sort_by = sort_by.lstrip('-')
        elif order == 'desc' and not sort_by.startswith('-'):
            sort_by = f'-{sort_by}'
        queryset = queryset.order_by(sort_by)

    # Pagination
    paginator = Paginator(queryset, per_page)
    try:
        page_obj = paginator.page(page)
    except PageNotAnInteger:
        page_obj = paginator.page(1)
    except EmptyPage:
        page_obj = paginator.page(paginator.num_pages)

    # Get log details for the current page only
    log_details = []
    for log in page_obj:
        detail = {
            'id': log.id,
            'timestamp': log.timestamp.strftime('%Y-%m-%d %H:%M:%S'),
            'device_identifier': log.device_identifier,
            'log_type': log.log_type,
            'description': log.description,
            'details': None
        }

        # Get specific log details based on type
        try:
            if log.log_type == 'activity':
                activity = log.activitylog
                detail['details'] = {
                    'window_title': activity.window_title,
                    'is_flagged': activity.is_flagged,
                    'has_screenshot': bool(activity.screenshot),
                    'screenshot_url': activity.screenshot.url if activity.screenshot else None,
                    'analysis': activity.analysis,
                    'keywords': activity.keywords or []
                }
            elif log.log_type == 'app_usage':
                app = log.appusagelog
                detail['details'] = {
                    'app_name': app.app_name,
                    'window_title': app.window_title,
                    'duration': app.duration,
                    'is_active': app.is_active
                }
            elif log.log_type == 'website_visit':
                website = log.websitevisitlog
                detail['details'] = {
                    'url': website.url,
                    'title': website.title,
                    'duration': website.duration
                }
            elif log.log_type == 'file_access':
                file = log.fileaccesslog
                detail['details'] = {
                    'file_path': file.file_path,
                    'operation': file.operation,
                    'process_name': file.process_name
                }
            elif log.log_type == 'usb_device':
                usb = log.usbdevicelog
                detail['details'] = {
                    'device_name': usb.device_name,
                    'vendor_id': usb.vendor_id,
                    'product_id': usb.product_id,
                    'serial_number': usb.serial_number,
                    'action': usb.action
                }
        except (ActivityLog.DoesNotExist, AppUsageLog.DoesNotExist,
                WebsiteVisitLog.DoesNotExist, FileAccessLog.DoesNotExist,
                USBDeviceLog.DoesNotExist):
            pass

        log_details.append(detail)

    context = {
        'logs': json.dumps(log_details, default=str),  # Serialize to JSON for JavaScript
        'page_obj': page_obj,
        'log_type': log_type,
        'date_from': date_from.strftime('%Y-%m-%d') if isinstance(date_from, datetime) else '',
        'date_to': date_to.strftime('%Y-%m-%d') if isinstance(date_to, datetime) else '',
        'keyword': keyword,
        'flagged_only': flagged_only,
        'has_screenshot': has_screenshot,
        'sort': sort_by,
        'order': order,
        'log_types': BaseLog.LOG_TYPES
    }
    return render(request, 'dashboard/logs_explorer.html', context)
