FROM php:8.2-fpm

RUN apt-get update && apt-get install -y curl unzip zip build-essential zlib1g-dev libzip-dev libpng-dev libjpeg-dev libjpeg62-turbo-dev libfreetype6-dev libwebp-dev libxpm-dev libgd-dev fontconfig locales 

RUN apt-get clean && rm -rf /var/lib/apt/lists/*

# install extensions
RUN docker-php-ext-configure gd --with-freetype --with-jpeg
RUN docker-php-ext-install pdo_mysql gd pdo zip

# install Composer
RUN curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer

WORKDIR /var/www/html

COPY composer.lock composer.json ./

RUN composer install --no-dev --optimize-autoloader

RUN chown -R www-data:www-data /var/www/html

EXPOSE 9000

CMD ["php-fpm"]