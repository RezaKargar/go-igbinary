FROM php:8.3-cli

# Install memcached extension dependencies
RUN apt-get update && apt-get install -y \
    libmemcached-dev \
    zlib1g-dev \
    && rm -rf /var/lib/apt/lists/*

# Install igbinary extension
RUN pecl install igbinary && docker-php-ext-enable igbinary

# Install memcached extension with igbinary support
RUN pecl install --configureoptions='enable-memcached-igbinary="yes"' memcached \
    && docker-php-ext-enable memcached

WORKDIR /app
COPY write_cache.php /app/write_cache.php
