# Laravel

## Project setup

Deployer detects PHP/Laravel projects by the presence of `composer.json`. Include a `Dockerfile` for the best results.

### Dockerfile

```dockerfile
FROM php:8.3-fpm-alpine

# Install extensions
RUN docker-php-ext-install pdo pdo_mysql pdo_pgsql opcache

# Install Composer
COPY --from=composer:2 /usr/composer /usr/bin/composer

WORKDIR /var/www/html
COPY composer.json composer.lock ./
RUN composer install --no-dev --optimize-autoloader --no-interaction

COPY . .

RUN php artisan config:cache \
    && php artisan route:cache \
    && php artisan view:cache

# Use the built-in server for simplicity
EXPOSE 8000
CMD ["php", "artisan", "serve", "--host=0.0.0.0", "--port=8000"]
```

For production, use nginx + PHP-FPM instead of `artisan serve`.

## Environment variables

Set your Laravel environment variables via the CLI:

```bash
deployer env set \
  APP_KEY=base64:your-app-key \
  APP_ENV=production \
  APP_DEBUG=false \
  LOG_CHANNEL=stderr
```

Do **not** include the `.env` file in your deployment archive. Use Deployer's environment variable management instead.

## Database setup

```bash
deployer db create postgres --name laravel-db
deployer db link <db-id> <app-id>
```

This sets `DATABASE_URL` on your app. Configure Laravel to use it in `config/database.php`:

```php
'default' => env('DB_CONNECTION', 'pgsql'),

'connections' => [
    'pgsql' => [
        'driver' => 'pgsql',
        'url' => env('DATABASE_URL'),
    ],
],
```

## Running migrations

Include migrations in your Dockerfile or run them after deployment:

```dockerfile
CMD php artisan migrate --force && php artisan serve --host=0.0.0.0 --port=8000
```

## Deploy

```bash
deployer init
deployer deploy
```

## Common issues

| Problem | Solution |
|---------|----------|
| `APP_KEY` missing | Generate with `php artisan key:generate --show` and set via `deployer env set` |
| Storage permissions | Run `chmod -R 775 storage bootstrap/cache` in Dockerfile |
| Composer memory error | Add `COMPOSER_MEMORY_LIMIT=-1` to your build |
