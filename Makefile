.PHONY: dev dev-up dev-down api web build clean monitoring-up monitoring-down docker-build k8s-apply k8s-delete docs docs-build

# Start infrastructure (PostgreSQL, Redis, Traefik, Registry)
dev-up:
	docker compose -f deployments/docker-compose.yml up -d

# Stop infrastructure
dev-down:
	docker compose -f deployments/docker-compose.yml down

# Run API server
api:
	cd api && go run ./cmd/server

# Run Vue.js dev server
web:
	cd web && npm run dev

# Run both API and Web (requires tmux or run in separate terminals)
dev: dev-up
	@echo "Infrastructure started. Run 'make api' and 'make web' in separate terminals."

# Build API binary
build-api:
	cd api && go build -o ../bin/api ./cmd/server

# Build Vue.js frontend
build-web:
	cd web && npm run build

# Build all
build: build-api build-web

# Clean build artifacts
clean:
	rm -rf bin/ web/dist/

# Start monitoring stack
monitoring-up:
	docker compose -f deployments/docker-compose.yml up -d prometheus grafana loki promtail alertmanager node-exporter postgres-exporter

# Stop monitoring stack
monitoring-down:
	docker compose -f deployments/docker-compose.yml stop prometheus grafana loki promtail alertmanager node-exporter postgres-exporter

# Run database migrations (placeholder)
migrate:
	@echo "Running migrations..."
	psql "$(DATABASE_URL)" -f scripts/init.sql

# Docker build
docker-build:
	docker build -t deployer-api:latest ./api
	docker build -t deployer-worker:latest -f ./api/Dockerfile.worker ./api
	docker build -t deployer-web:latest ./web
	docker build -t deployer-cli:latest ./cli

# Kubernetes deploy
k8s-apply:
	kubectl apply -k deployments/k8s/

k8s-delete:
	kubectl delete -k deployments/k8s/

# Documentation
docs:
	cd docs && npm run docs:dev

docs-build:
	cd docs && npm run docs:build
