from django.urls import path, include
from rest_framework.routers import DefaultRouter
from .views import (
    ActivityLogViewSet,
    AppUsageLogViewSet,
    WebsiteVisitLogViewSet,
    FileAccessLogViewSet,
    USBDeviceLogViewSet,
    BulkMonitoringViewSet
)

router = DefaultRouter()
router.register(r'logs', ActivityLogViewSet)
router.register(r'app-usage', AppUsageLogViewSet)
router.register(r'website-visits', WebsiteVisitLogViewSet)
router.register(r'file-access', FileAccessLogViewSet)
router.register(r'usb-devices', USBDeviceLogViewSet)
router.register(r'bulk', BulkMonitoringViewSet, basename='bulk')

urlpatterns = [
    path('', include(router.urls)),
] 