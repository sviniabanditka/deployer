# Python / Django / Flask / FastAPI

## Project setup

Deployer detects Python projects by the presence of `requirements.txt`.

### requirements.txt

```
flask==3.0.0
gunicorn==21.2.0
```

Or for Django:

```
django==5.0
gunicorn==21.2.0
psycopg2-binary==2.9.9
```

### Procfile

Create a `Procfile` in your project root to specify the start command:

```
web: gunicorn app:app --bind 0.0.0.0:$PORT
```

Django:

```
web: gunicorn myproject.wsgi:application --bind 0.0.0.0:$PORT
```

FastAPI:

```
web: uvicorn main:app --host 0.0.0.0 --port $PORT
```

## Flask example

```python
import os
from flask import Flask

app = Flask(__name__)

@app.route('/')
def hello():
    return {'message': 'Hello from Deployer!'}

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 5000))
    app.run(host='0.0.0.0', port=port)
```

## Django example

In `settings.py`:

```python
import os
import dj_database_url

ALLOWED_HOSTS = ['*']
SECRET_KEY = os.environ.get('SECRET_KEY', 'change-me')

DATABASES = {
    'default': dj_database_url.config(default=os.environ.get('DATABASE_URL'))
}

STATIC_ROOT = os.path.join(BASE_DIR, 'staticfiles')
```

## Deploy

```bash
deployer init
deployer deploy
```

## Using a virtual environment

Deployer installs dependencies from `requirements.txt` during the build. You do not need to include your virtual environment in the upload. Make sure `venv/` or `.venv/` is in `.deployerignore`.

## Using a custom Dockerfile

```dockerfile
FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE 8000
CMD ["gunicorn", "app:app", "--bind", "0.0.0.0:8000"]
```

## Adding a database

```bash
deployer db create postgres
deployer db link <db-id> <app-id>
deployer env set SECRET_KEY=your-secret-key
```

## Common issues

| Problem | Solution |
|---------|----------|
| `ModuleNotFoundError` | Ensure the module is in `requirements.txt` |
| Static files not loading | Run `collectstatic` in the build or Dockerfile |
| Timeout on start | Use gunicorn or uvicorn, not the dev server |
