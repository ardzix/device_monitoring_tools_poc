# Database Configuration
USE_POSTGRES = os.getenv('USE_POSTGRES', 'False').lower() == 'true'

if USE_POSTGRES:
    DATABASES = {
        'default': {
            'ENGINE': 'django.db.backends.postgresql',
            'NAME': os.getenv('POSTGRES_DB', 'employeemonitoring'),
            'USER': os.getenv('POSTGRES_USER', 'postgres'),
            'PASSWORD': os.getenv('POSTGRES_PASSWORD', ''),
            'HOST': os.getenv('POSTGRES_HOST', 'localhost'),
            'PORT': os.getenv('POSTGRES_PORT', '5432'),
        }
    }
else:
    DATABASES = {
        'default': {
            'ENGINE': 'django.db.backends.sqlite3',
            'NAME': BASE_DIR / os.getenv('SQLITE_NAME', 'db.sqlite3'),
        }
    }

# Storage Configuration
USE_S3 = os.getenv('USE_S3', 'False').lower() == 'true'

if USE_S3:
    # S3 Storage Settings
    AWS_ACCESS_KEY_ID = os.getenv('AWS_ACCESS_KEY_ID')
    AWS_SECRET_ACCESS_KEY = os.getenv('AWS_SECRET_ACCESS_KEY')
    AWS_STORAGE_BUCKET_NAME = os.getenv('AWS_STORAGE_BUCKET_NAME')
    AWS_S3_ENDPOINT_URL = os.getenv('AWS_S3_ENDPOINT_URL')
    AWS_S3_REGION_NAME = os.getenv('AWS_S3_REGION_NAME')
    AWS_DEFAULT_ACL = os.getenv('AWS_DEFAULT_ACL', 'private')
    AWS_S3_FILE_OVERWRITE = os.getenv('AWS_S3_FILE_OVERWRITE', 'False').lower() == 'true'
    AWS_QUERYSTRING_AUTH = os.getenv('AWS_QUERYSTRING_AUTH', 'True').lower() == 'true'

    # S3 Storage Backend Configuration
    DEFAULT_FILE_STORAGE = 'storages.backends.s3boto3.S3Boto3Storage'
    STATICFILES_STORAGE = 'storages.backends.s3boto3.S3Boto3Storage'
else:
    # Local Storage Settings
    MEDIA_ROOT = os.path.join(BASE_DIR, os.getenv('MEDIA_ROOT', 'media'))
    STATIC_ROOT = os.path.join(BASE_DIR, 'static')
    SCREENSHOTS_DIR = os.path.join(MEDIA_ROOT, os.getenv('SCREENSHOTS_DIR', 'screenshots'))
    LOGS_DIR = os.path.join(BASE_DIR, os.getenv('LOGS_DIR', 'logs'))

    # Create necessary directories
    for directory in [MEDIA_ROOT, STATIC_ROOT, SCREENSHOTS_DIR, LOGS_DIR]:
        os.makedirs(directory, exist_ok=True) 