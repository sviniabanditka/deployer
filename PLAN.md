# Deployer — PaaS-платформа для простого деплоя приложений

> Стек: Go + Vue.js

---

## Архитектура

```
┌─────────────┐     ┌──────────────┐     ┌──────────────────┐
│  Frontend   │────▶│  API Gateway │────▶│  Core Services    │
│  (Vue.js)   │     │  (REST/WS)   │     │                   │
└─────────────┘     └──────────────┘     │  ┌──────────────┐ │
                                         │  │ Auth Service  │ │
┌─────────────┐                          │  ├──────────────┤ │
│  CLI (Go)   │──────────────────────────│  │ Build Service │ │
└─────────────┘                          │  ├──────────────┤ │
                                         │  │ Deploy Svc   │ │
┌─────────────┐                          │  ├──────────────┤ │
│  Git Push   │──────────────────────────│  │ DB Manager   │ │
│  (webhook)  │                          │  ├──────────────┤ │
└─────────────┘                          │  │ Billing Svc  │ │
                                         │  ├──────────────┤ │
                                         │  │ Proxy/Router │ │
                                         │  └──────────────┘ │
                                         └────────┬──────────┘
                                                  │
                                         ┌────────▼──────────┐
                                         │  Docker Engine /   │
                                         │  Kubernetes        │
                                         └───────────────────┘
```

---

## Стек технологий

| Компонент       | Технология                              |
|-----------------|------------------------------------------|
| API Backend     | Go (GoFiber)                             |
| Frontend        | Vue.js 3 + Vite + Tailwind CSS           |
| CLI             | Go (Cobra)                               |
| Оркестрация     | Docker Swarm (MVP) → Kubernetes (scale)  |
| Reverse Proxy   | Traefik (авто-SSL, авто-роутинг)         |
| Сборка образов  | Nixpacks / Cloud Native Buildpacks       |
| БД платформы    | PostgreSQL                               |
| Очереди/кеш     | Redis + Asynq                            |
| Хранилище       | S3 (MinIO self-hosted или AWS S3)        |
| Мониторинг      | Prometheus + Grafana                     |
| Логи            | Loki                                     |
| Платежи         | Stripe                                   |
| Хостинг         | EU-дата-центры (Hetzner / OVH / AWS EU)  |

---

## Структура проекта

```
deployer/
├── api/                  # Go API сервер
│   ├── cmd/              # точка входа
│   ├── internal/
│   │   ├── handler/      # HTTP handlers
│   │   ├── service/      # бизнес-логика
│   │   │   ├── auth/
│   │   │   ├── build/
│   │   │   ├── deploy/
│   │   │   ├── database/
│   │   │   └── billing/
│   │   ├── model/
│   │   ├── repository/
│   │   └── middleware/
│   ├── pkg/              # переиспользуемые пакеты
│   └── go.mod
├── web/                  # Vue.js frontend
│   ├── src/
│   │   ├── views/
│   │   ├── components/
│   │   ├── composables/
│   │   ├── stores/       # Pinia
│   │   ├── router/
│   │   └── api/          # API client
│   ├── package.json
│   └── vite.config.ts
├── cli/                  # CLI утилита (Go + Cobra)
│   ├── cmd/
│   └── go.mod
├── builder/              # Build worker (Nixpacks/Buildpacks)
├── proxy/                # Конфигурация Traefik
├── scripts/              # Миграции, скрипты установки
├── deployments/
│   ├── docker-compose.yml
│   └── k8s/              # Kubernetes манифесты
├── PLAN.md
└── README.md
```

---

## Фазы разработки

### Фаза 1 — Ядро / MVP

> Цель: загрузить ZIP или Dockerfile → приложение работает на поддомене с HTTPS.

- [x] **1.1. Инфраструктура разработки** ✅
  - docker-compose: Traefik + PostgreSQL + Redis + Docker Registry
  - Makefile для удобной сборки и запуска
  - Базовая конфигурация (env-файлы, config struct)
  - SQL-схема БД (users, apps, deployments, env_vars)

- [x] **1.2. Auth Service** ✅
  - Регистрация / логин (email + password, bcrypt)
  - JWT access (15min) + refresh (7d) tokens
  - JWT middleware для защищённых маршрутов
  - OAuth: GitHub, GitLab — TODO

- [x] **1.3. API — управление приложениями** ✅
  - CRUD приложений (create, list, get, delete)
  - Управление переменными окружения (env vars)
  - Загрузка исходного кода (ZIP upload)
  - WebSocket endpoint для стриминга логов

- [x] **1.4. Build Service** ✅
  - Worker на Go + Asynq (Redis queue, concurrency 5)
  - Приём ZIP → распаковка → определение рантайма (auto-detect)
  - Сборка Docker-образа через Nixpacks или Dockerfile
  - Поддержка: Node.js, Python, Go, Ruby, Java, Rust, .NET, PHP
  - Push образов в local Docker Registry

- [x] **1.5. Deploy Service** ✅
  - Запуск контейнера через Docker API
  - Health-checks, restart policy (unless-stopped)
  - Управление ресурсами (CPU/RAM лимиты)
  - Привязка поддомена `{app-slug}.localhost` (configurable domain)
  - Start/Stop/Remove контейнеров
  - Стриминг логов и статистика (CPU/RAM/Network)

- [x] **1.6. Reverse Proxy (Traefik)** ✅
  - Динамическая маршрутизация по поддоменам (Docker labels)
  - Автоматические SSL-сертификаты (Let's Encrypt)
  - Load balancing

- [x] **1.7. Web Dashboard (Vue.js) — минимальный** ✅
  - Страницы: логин, регистрация, dashboard, список приложений, деталь приложения
  - Загрузка ZIP через drag-n-drop
  - Просмотр логов в реальном времени (WebSocket)
  - Управление env vars
  - Мониторинг статуса приложений

---

### Фаза 2 — Git-деплой и UX

- [x] **2.1. Git Integration** ✅
  - Подключение GitHub/GitLab репозитория (OAuth + webhooks)
  - Автодеплой при push в настраиваемую ветку
  - Верификация webhook-подписей (HMAC SHA-256 / token)
  - Создание/удаление webhooks через API провайдеров
  - Preview deployments для pull requests — TODO

- [x] **2.2. CLI-утилита** ✅
  ```bash
  deployer login / register      # авторизация
  deployer init                  # инициализация проекта
  deployer deploy                # деплой из текущей папки (ZIP)
  deployer logs -f               # стриминг логов (WebSocket)
  deployer env list/set/unset    # переменные окружения
  deployer apps list/info/delete # управление приложениями
  deployer start / stop          # запуск/остановка
  deployer scale web=3           # масштабирование (placeholder)
  ```

- [x] **2.3. Dashboard — расширение** ✅
  - 6-табный AppDetail: Overview, Deployments, Env Vars, Git, Logs, Settings
  - Подключение Git-репозитория из UI (GitHub/GitLab)
  - Мониторинг ресурсов (CPU/RAM/сеть) — бары с авто-обновлением
  - LogViewer — терминал с WebSocket стримингом
  - EnvEditor — редактор с маскировкой значений
  - DeploymentList — история деплоев с build log
  - StatusBadge, ConfirmModal — переиспользуемые компоненты
  - Управление custom domains — TODO
  - Blue-green / rolling деплой (zero-downtime) — TODO

---

### Фаза 3 — Managed Databases

- [x] **3.1. Database as a Service** ✅
  - PostgreSQL, MySQL, MongoDB, Redis — как managed Docker-контейнеры
  - Создание одной кнопкой (UI) / одной командой (`deployer db create postgres`)
  - Автоматическое подключение через `DATABASE_URL` env var (link/unlink)
  - Бэкапы: создание (pg_dump/mysqldump/mongodump), восстановление, список
  - Генерация безопасных credentials (crypto/rand)
  - API: 11 endpoints (CRUD, start/stop, link/unlink, backup/restore)
  - CLI: `deployer db create/list/info/delete/stop/start/link/unlink/backup/backups/restore`
  - UI: DatabaseListView, DatabaseDetailView (3 таба), ConnectionInfo компонент
  - Бэкапы по расписанию (S3 storage) — TODO
  - Web-интерфейс для просмотра данных (pgAdmin/Adminer) — TODO

---

### Фаза 4 — Биллинг

- [x] **4.1. Тарифные планы** ✅
  | План       | Цена       | Приложения | RAM   | vCPU | Особенности           |
  |------------|------------|------------|-------|------|-----------------------|
  | Free       | €0         | 1          | 512MB | 0.5  | Поддомен              |
  | Starter    | €5/мес     | 3          | 1GB   | 1    | Custom domain, 1 DB   |
  | Pro        | €19/мес    | 10         | 4GB   | 2    | 3 DB, приоритет       |
  | Business   | €49/мес    | Unlimited  | 16GB  | 8    | Unlimited DB, SLA     |
  - Планы сидированы в БД, модель Plan с лимитами

- [x] **4.2. Stripe интеграция** ✅
  - Подписки через Stripe SDK (stripe-go v82)
  - Stripe Checkout для оплаты
  - Stripe Billing Portal для самообслуживания
  - Webhook-обработчик: invoice.paid, invoice.payment_failed, subscription.updated/deleted
  - Invoices с историей платежей
  - EU VAT handling — TODO (через Stripe Tax)

- [x] **4.3. Квоты и лимиты** ✅
  - Enforcement лимитов по тарифу (EnforceAppQuota, EnforceDBQuota)
  - 403 при превышении квоты на создание app/db
  - UsageSummary: текущее потребление vs лимиты
  - UI: UsageBar с цветовой индикацией (green/yellow/red)
  - Rate limiting API ✅ (Redis sliding window, 100/min public, 1000/min auth)

---

### Фаза 5 — Production-ready

- [x] **5.1. Мониторинг и наблюдаемость** ✅
  - Prometheus + Grafana для метрик платформы (10-панельный dashboard)
  - Prometheus middleware: http_requests_total, http_request_duration_seconds, active_connections
  - Loki + Promtail для сбора Docker-логов (30 дней retention)
  - Alertmanager: 8 правил алертов (AppDown, HighCPU, HighMemory, DiskSpace, Latency, ErrorRate, PostgresDown, BuildBacklog)
  - Node-exporter + Postgres-exporter
  - Rate limiting middleware (Redis sliding window)
  - Статус-страница (uptime monitoring) — TODO

- [x] **5.2. Безопасность** ✅
  - Security headers middleware (HSTS, CSP, X-Frame-Options, etc.)
  - Сканирование образов (Trivy) — ScanImage service
  - 2FA для пользователей (TOTP) — enable/verify/validate/disable
  - Account lockout: 5 неудачных попыток → блокировка 15 мин
  - GDPR compliance: data export (JSON), account deletion, anonymization
  - Settings UI: профиль, безопасность (2FA), данные и приватность
  - Login UI: поддержка 2FA flow с 6-digit input
  - Изоляция сети между контейнерами — TODO (Docker network policies)
  - SOC 2 preparation — TODO

- [x] **5.3. Масштабирование** ✅
  - Kubernetes манифесты: 13 файлов (namespace, configmap, secrets, postgres, redis, registry, api, worker, web, ingress, monitoring, cert-manager, kustomization)
  - HPA: API (2-10 pods, 70% CPU), Worker (2-8 pods, 80% CPU)
  - Dockerfiles: api, worker (с nixpacks+docker-cli), web (nginx), cli
  - nginx.conf: SPA routing, gzip, кеширование статики 1 год
  - cert-manager: Let's Encrypt (staging + prod), wildcard *.apps.deployer.dev
  - Ingress: api.deployer.dev, app.deployer.dev, *.apps.deployer.dev
  - Multi-region (EU West + EU Central) — TODO (infrastructure provisioning)

- [x] **5.4. Документация и маркетинг** ✅
  - VitePress документация: 31 страница
  - Landing page с hero + 6 features
  - Guide: getting started, quick start, deploy (ZIP/Git/Dockerfile), env vars, custom domains
  - Database guides: PostgreSQL, MySQL, MongoDB, Redis, backups
  - Framework guides: Node.js, Python, Go, Next.js, Laravel, Static
  - API Reference: Auth, Apps, Deployments, Databases, Billing (с curl примерами)
  - CLI Reference: installation + all commands
  - Pricing page
  - Blog — TODO

---

## Поддерживаемые рантаймы (через Nixpacks)

- Node.js (Express, Next.js, Nuxt, Nest.js)
- Python (Django, Flask, FastAPI)
- Go (любой фреймворк)
- .NET Core
- PHP (Laravel, Symfony)
- Ruby (Rails)
- Java (Spring Boot)
- Rust
- Static sites (HTML/CSS/JS)
- Произвольный Dockerfile

---

## EU-специфика

- Дата-центры только в EU (Hetzner Falkenstein/Helsinki, OVH, AWS eu-central-1)
- GDPR compliance by design
- Stripe с EU VAT поддержкой
- Интерфейс: English (primary), возможно DE/FR позже
- Юридическое лицо в EU
- DPA (Data Processing Agreement) для бизнес-клиентов
- Cookie consent banner
