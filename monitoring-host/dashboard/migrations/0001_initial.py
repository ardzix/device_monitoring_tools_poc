# Generated by Django 5.2 on 2025-04-21 07:39

import django.db.models.deletion
from django.db import migrations, models


class Migration(migrations.Migration):

    initial = True

    dependencies = [
    ]

    operations = [
        migrations.CreateModel(
            name='BaseLog',
            fields=[
                ('id', models.BigAutoField(auto_created=True, primary_key=True, serialize=False, verbose_name='ID')),
                ('timestamp', models.DateTimeField(auto_now_add=True)),
                ('description', models.TextField(blank=True, null=True)),
                ('log_type', models.CharField(choices=[('activity', 'Activity'), ('app_usage', 'App Usage'), ('website_visit', 'Website Visit'), ('file_access', 'File Access'), ('usb_device', 'USB Device')], max_length=20)),
                ('device_identifier', models.CharField(help_text='Unique identifier for the device', max_length=255)),
            ],
            options={
                'verbose_name': 'Base Log',
                'verbose_name_plural': 'Base Logs',
                'ordering': ['-timestamp'],
            },
        ),
        migrations.CreateModel(
            name='ActivityLog',
            fields=[
                ('baselog_ptr', models.OneToOneField(auto_created=True, on_delete=django.db.models.deletion.CASCADE, parent_link=True, primary_key=True, serialize=False, to='dashboard.baselog')),
                ('window_title', models.CharField(max_length=255)),
                ('clipboard', models.TextField(blank=True)),
                ('screenshot', models.ImageField(blank=True, null=True, upload_to='screenshots/')),
                ('is_flagged', models.BooleanField(default=False)),
                ('confidence', models.FloatField(default=0.0)),
                ('analysis', models.TextField(blank=True)),
                ('keywords', models.JSONField(blank=True, default=list, null=True)),
            ],
            bases=('dashboard.baselog',),
        ),
        migrations.CreateModel(
            name='AppUsageLog',
            fields=[
                ('baselog_ptr', models.OneToOneField(auto_created=True, on_delete=django.db.models.deletion.CASCADE, parent_link=True, primary_key=True, serialize=False, to='dashboard.baselog')),
                ('app_name', models.CharField(max_length=255)),
                ('window_title', models.CharField(max_length=255)),
                ('duration', models.IntegerField(default=0)),
                ('is_active', models.BooleanField(default=False)),
            ],
            bases=('dashboard.baselog',),
        ),
        migrations.CreateModel(
            name='FileAccessLog',
            fields=[
                ('baselog_ptr', models.OneToOneField(auto_created=True, on_delete=django.db.models.deletion.CASCADE, parent_link=True, primary_key=True, serialize=False, to='dashboard.baselog')),
                ('file_path', models.CharField(max_length=512)),
                ('operation', models.CharField(max_length=50)),
                ('process_name', models.CharField(max_length=255)),
            ],
            bases=('dashboard.baselog',),
        ),
        migrations.CreateModel(
            name='USBDeviceLog',
            fields=[
                ('baselog_ptr', models.OneToOneField(auto_created=True, on_delete=django.db.models.deletion.CASCADE, parent_link=True, primary_key=True, serialize=False, to='dashboard.baselog')),
                ('device_name', models.CharField(max_length=255)),
                ('vendor_id', models.CharField(max_length=10)),
                ('product_id', models.CharField(max_length=10)),
                ('serial_number', models.CharField(blank=True, max_length=255)),
                ('action', models.CharField(max_length=50)),
            ],
            bases=('dashboard.baselog',),
        ),
        migrations.CreateModel(
            name='WebsiteVisitLog',
            fields=[
                ('baselog_ptr', models.OneToOneField(auto_created=True, on_delete=django.db.models.deletion.CASCADE, parent_link=True, primary_key=True, serialize=False, to='dashboard.baselog')),
                ('url', models.URLField()),
                ('title', models.CharField(max_length=255)),
                ('duration', models.IntegerField(default=0)),
            ],
            bases=('dashboard.baselog',),
        ),
    ]
