CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE apps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'created',
    runtime VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    version INTEGER NOT NULL DEFAULT 1,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    image_tag VARCHAR(255),
    build_log TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE env_vars (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(app_id, key)
);

CREATE INDEX idx_apps_user_id ON apps(user_id);
CREATE INDEX idx_apps_slug ON apps(slug);
CREATE INDEX idx_deployments_app_id ON deployments(app_id);
CREATE INDEX idx_env_vars_app_id ON env_vars(app_id);

CREATE TABLE git_connections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    repo_url VARCHAR(500) NOT NULL,
    repo_owner VARCHAR(255) NOT NULL,
    repo_name VARCHAR(255) NOT NULL,
    branch VARCHAR(255) NOT NULL DEFAULT 'main',
    auto_deploy BOOLEAN NOT NULL DEFAULT true,
    webhook_secret VARCHAR(255) NOT NULL,
    access_token VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(app_id)
);
CREATE INDEX idx_git_connections_app_id ON git_connections(app_id);

CREATE TABLE managed_databases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    app_id UUID REFERENCES apps(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    engine VARCHAR(50) NOT NULL, -- 'postgres', 'mysql', 'mongodb', 'redis'
    version VARCHAR(20) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'creating', -- creating, running, stopped, failed, deleted
    host VARCHAR(255),
    port INTEGER,
    db_name VARCHAR(255),
    username VARCHAR(255),
    password VARCHAR(255),
    connection_url TEXT,
    container_id VARCHAR(255),
    memory_limit BIGINT NOT NULL DEFAULT 268435456, -- 256MB
    storage_used BIGINT NOT NULL DEFAULT 0,
    storage_limit BIGINT NOT NULL DEFAULT 1073741824, -- 1GB
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_managed_databases_user_id ON managed_databases(user_id);
CREATE INDEX idx_managed_databases_app_id ON managed_databases(app_id);

CREATE TABLE database_backups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    database_id UUID NOT NULL REFERENCES managed_databases(id) ON DELETE CASCADE,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'creating', -- creating, completed, failed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_database_backups_database_id ON database_backups(database_id);

-- Billing plans
CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    price_cents INTEGER NOT NULL DEFAULT 0, -- monthly price in EUR cents
    app_limit INTEGER NOT NULL DEFAULT 1,
    db_limit INTEGER NOT NULL DEFAULT 0,
    memory_limit BIGINT NOT NULL DEFAULT 536870912, -- 512MB per app
    cpu_limit FLOAT NOT NULL DEFAULT 0.5,
    storage_limit BIGINT NOT NULL DEFAULT 1073741824, -- 1GB
    custom_domains BOOLEAN NOT NULL DEFAULT false,
    priority_support BOOLEAN NOT NULL DEFAULT false,
    stripe_price_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Seed default plans
INSERT INTO plans (name, display_name, price_cents, app_limit, db_limit, memory_limit, cpu_limit, storage_limit, custom_domains, priority_support) VALUES
('free', 'Free', 0, 1, 0, 536870912, 0.5, 1073741824, false, false),
('starter', 'Starter', 500, 3, 1, 1073741824, 1.0, 5368709120, true, false),
('pro', 'Pro', 1900, 10, 3, 4294967296, 2.0, 21474836480, true, false),
('business', 'Business', 4900, -1, -1, 17179869184, 8.0, 107374182400, true, true);

-- User subscriptions
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES plans(id),
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, canceled, past_due, trialing
    stripe_customer_id VARCHAR(255),
    stripe_subscription_id VARCHAR(255),
    current_period_start TIMESTAMP WITH TIME ZONE,
    current_period_end TIMESTAMP WITH TIME ZONE,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id)
);
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_stripe_customer_id ON subscriptions(stripe_customer_id);

-- Invoices
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stripe_invoice_id VARCHAR(255),
    amount_cents INTEGER NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'eur',
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, paid, failed
    description TEXT,
    invoice_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_invoices_user_id ON invoices(user_id);

-- Usage tracking
CREATE TABLE usage_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    resource_type VARCHAR(50) NOT NULL, -- 'app', 'database', 'bandwidth', 'storage'
    resource_id UUID,
    quantity BIGINT NOT NULL DEFAULT 0,
    unit VARCHAR(50) NOT NULL, -- 'minutes', 'bytes', 'requests'
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_usage_records_user_id ON usage_records(user_id);
CREATE INDEX idx_usage_records_recorded_at ON usage_records(recorded_at);

-- Security: 2FA and account locking columns
ALTER TABLE users ADD COLUMN IF NOT EXISTS two_factor_secret VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS two_factor_enabled BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS login_attempts INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP WITH TIME ZONE;

-- Preview deployments
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS is_preview BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS preview_url VARCHAR(500);
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS pull_request_number INTEGER;
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS pull_request_url VARCHAR(500);

-- Incidents table for status page
CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    component VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    severity VARCHAR(50) NOT NULL DEFAULT 'minor', -- minor, major, critical
    status VARCHAR(50) NOT NULL DEFAULT 'investigating', -- investigating, identified, monitoring, resolved
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_incidents_started_at ON incidents(started_at);

-- OAuth support for users
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_id VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_oauth ON users(oauth_provider, oauth_id) WHERE oauth_provider IS NOT NULL;

-- Custom domains
CREATE TABLE IF NOT EXISTS custom_domains (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    domain VARCHAR(255) NOT NULL UNIQUE,
    verification_token VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending_verification',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_custom_domains_app_id ON custom_domains(app_id);

-- Billing addresses for EU VAT support
CREATE TABLE IF NOT EXISTS billing_addresses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    country VARCHAR(10) NOT NULL,
    postal_code VARCHAR(20) NOT NULL DEFAULT '',
    vat_number VARCHAR(50) NOT NULL DEFAULT '',
    tax_exempt BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id)
);
CREATE INDEX IF NOT EXISTS idx_billing_addresses_user_id ON billing_addresses(user_id);

-- Scheduled database backup settings
ALTER TABLE managed_databases ADD COLUMN IF NOT EXISTS auto_backup BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE managed_databases ADD COLUMN IF NOT EXISTS backup_retention INTEGER NOT NULL DEFAULT 7;
ALTER TABLE managed_databases ADD COLUMN IF NOT EXISTS backup_schedule VARCHAR(100) NOT NULL DEFAULT '0 2 * * *';

-- User network isolation
CREATE TABLE IF NOT EXISTS user_networks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    network_id VARCHAR(255) NOT NULL,
    network_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id)
);
