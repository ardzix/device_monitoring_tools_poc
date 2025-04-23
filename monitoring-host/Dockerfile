# Use an official Python runtime as a parent image
FROM python:3.11-slim-buster

# Set the working directory in the container
WORKDIR /usr/src/app

# Install system dependencies for building Python libraries
RUN apt-get update && apt-get install -y \
    build-essential \
    python3-dev \
    libpq-dev \
    libpcre3 \
    libpcre3-dev \
    libssl-dev \
    libffi-dev \
    supervisor \
    curl \
    && apt-get clean && rm -rf /var/lib/apt/lists/* 

# Create necessary directories
RUN mkdir -p /usr/src/app/monitoring-host

# Copy the requirements file into the container
COPY monitoring-host/requirements.txt /usr/src/app/monitoring-host/
RUN ls -la /usr/src/app/monitoring-host/requirements.txt

# Install Python dependencies with binary wheels
RUN pip install --no-cache-dir -r /usr/src/app/monitoring-host/requirements.txt

# Copy the application code into the container
COPY monitoring-host/ /usr/src/app/monitoring-host/
RUN ls -la /usr/src/app/monitoring-host/

# Set environment variables
ENV DJANGO_SETTINGS_MODULE=monitoring_host.settings
ENV PYTHONUNBUFFERED=1
ENV PYTHONPATH=/usr/src/app/monitoring-host

# Collect static files
RUN cd /usr/src/app/monitoring-host && python manage.py collectstatic --noinput

# Expose port for uWSGI/Django
EXPOSE 8000

# Create supervisor directory and copy configuration
RUN mkdir -p /etc/supervisor/conf.d
COPY monitoring-host/supervisord.conf /etc/supervisor/conf.d/supervisord.conf
RUN ls -la /etc/supervisor/conf.d/supervisord.conf

# Run Supervisor to manage processes
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"] 