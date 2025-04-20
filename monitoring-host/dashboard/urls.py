from django.urls import path, include
from rest_framework.routers import DefaultRouter
from .views import (
    ActivityLogViewSet,
    AppUsageLogViewSet,
    WebsiteVisitLogViewSet,
    FileAccessLogViewSet,
    USBDeviceLogViewSet,
    BulkMonitoringViewSet,
    dashboard_view,
    logs_explorer_view
)

router = DefaultRouter()
router.register(r'logs', ActivityLogViewSet)
router.register(r'app-usage', AppUsageLogViewSet)
router.register(r'website-visits', WebsiteVisitLogViewSet)
router.register(r'file-access', FileAccessLogViewSet)
router.register(r'usb-devices', USBDeviceLogViewSet)
router.register(r'bulk', BulkMonitoringViewSet, basename='bulk')

urlpatterns = [
    path('dashboard/', dashboard_view, name='dashboard'),
    path('api/', include(router.urls)),
    path('logs/', logs_explorer_view, name='logs_explorer'),
] 