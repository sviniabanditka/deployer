package config

import (
	"os"
	"strconv"
	"time"
)

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type Config struct {
	ServerPort   string
	DatabaseURL  string
	RedisURL     string
	JWTSecret    string
	JWTExpiry    time.Duration
	RegistryURL        string
	UploadDir          string
	NixpacksPath       string
	AppDomain          string
	DockerNetwork      string
	DefaultMemoryLimit int64
	DefaultCPULimit    float64
	WebhookBaseURL     string
	GitHubClientID     string
	GitHubClientSecret string
	GitLabClientID     string
	GitLabClientSecret string
	BackupDir              string
	DefaultDBMemoryLimit   int64
	DefaultDBStorageLimit  int64
	StripeSecretKey            string
	StripeWebhookSecret        string
	BillingPortalReturnURL     string
	GitHubRedirectURL          string
	GitLabRedirectURL          string
	OAuthFrontendRedirectURL   string
}

func Load() *Config {
	expiry := 15 * time.Minute
	if v := os.Getenv("JWT_EXPIRY_MINUTES"); v != "" {
		if mins, err := strconv.Atoi(v); err == nil {
			expiry = time.Duration(mins) * time.Minute
		}
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/deployer?sslmode=disable"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production"
	}

	registryURL := os.Getenv("REGISTRY_URL")
	if registryURL == "" {
		registryURL = "localhost:5000"
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	nixpacksPath := os.Getenv("NIXPACKS_PATH")
	if nixpacksPath == "" {
		nixpacksPath = "nixpacks"
	}

	appDomain := envOrDefault("APP_DOMAIN", "localhost")
	dockerNetwork := envOrDefault("DOCKER_NETWORK", "deployer")

	var defaultMemoryLimit int64 = 536870912 // 512MB
	if v := os.Getenv("DEFAULT_MEMORY_LIMIT"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			defaultMemoryLimit = n
		}
	}

	defaultCPULimit := 0.5
	if v := os.Getenv("DEFAULT_CPU_LIMIT"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			defaultCPULimit = f
		}
	}

	webhookBaseURL := envOrDefault("WEBHOOK_BASE_URL", "http://localhost:3000")
	gitHubClientID := os.Getenv("GITHUB_CLIENT_ID")
	gitHubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	gitLabClientID := os.Getenv("GITLAB_CLIENT_ID")
	gitLabClientSecret := os.Getenv("GITLAB_CLIENT_SECRET")

	backupDir := envOrDefault("BACKUP_DIR", "./backups")

	var defaultDBMemoryLimit int64 = 268435456 // 256MB
	if v := os.Getenv("DEFAULT_DB_MEMORY_LIMIT"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			defaultDBMemoryLimit = n
		}
	}

	var defaultDBStorageLimit int64 = 1073741824 // 1GB
	if v := os.Getenv("DEFAULT_DB_STORAGE_LIMIT"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			defaultDBStorageLimit = n
		}
	}

	stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
	stripeWebhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	billingPortalReturnURL := envOrDefault("BILLING_PORTAL_RETURN_URL", "http://localhost:5173/billing")

	gitHubRedirectURL := envOrDefault("GITHUB_REDIRECT_URL", "http://localhost:8080/api/v1/auth/github/callback")
	gitLabRedirectURL := envOrDefault("GITLAB_REDIRECT_URL", "http://localhost:8080/api/v1/auth/gitlab/callback")
	oauthFrontendRedirectURL := envOrDefault("OAUTH_FRONTEND_REDIRECT_URL", "http://localhost:5173/auth/callback")

	return &Config{
		ServerPort:             port,
		DatabaseURL:            dbURL,
		RedisURL:               redisURL,
		JWTSecret:              jwtSecret,
		JWTExpiry:              expiry,
		RegistryURL:            registryURL,
		UploadDir:              uploadDir,
		NixpacksPath:           nixpacksPath,
		AppDomain:              appDomain,
		DockerNetwork:          dockerNetwork,
		DefaultMemoryLimit:     defaultMemoryLimit,
		DefaultCPULimit:        defaultCPULimit,
		WebhookBaseURL:         webhookBaseURL,
		GitHubClientID:         gitHubClientID,
		GitHubClientSecret:     gitHubClientSecret,
		GitLabClientID:         gitLabClientID,
		GitLabClientSecret:     gitLabClientSecret,
		BackupDir:              backupDir,
		DefaultDBMemoryLimit:   defaultDBMemoryLimit,
		DefaultDBStorageLimit:  defaultDBStorageLimit,
		StripeSecretKey:          stripeSecretKey,
		StripeWebhookSecret:      stripeWebhookSecret,
		BillingPortalReturnURL:   billingPortalReturnURL,
		GitHubRedirectURL:        gitHubRedirectURL,
		GitLabRedirectURL:        gitLabRedirectURL,
		OAuthFrontendRedirectURL: oauthFrontendRedirectURL,
	}
}
