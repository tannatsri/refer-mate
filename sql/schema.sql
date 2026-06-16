CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    google_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    profile_picture TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS oauth_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    provider VARCHAR(50),
    access_token TEXT,
    refresh_token TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS email_templates (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255),
    subject TEXT,
    body TEXT,
    variables JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS campaigns (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    template_id BIGINT REFERENCES email_templates(id),
    campaign_name VARCHAR(255),
    description TEXT,
    status VARCHAR(50) DEFAULT 'DRAFT',
    scheduled_at TIMESTAMP,
    total_recipients INT DEFAULT 0,
    sent_count INT DEFAULT 0,
    opened_count INT DEFAULT 0,
    clicked_count INT DEFAULT 0,
    replied_count INT DEFAULT 0,
    failed_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS campaign_recipients (
    id BIGSERIAL PRIMARY KEY,
    campaign_id BIGINT REFERENCES campaigns(id) ON DELETE CASCADE,
    recipient_name VARCHAR(255),
    recipient_email VARCHAR(255),
    company VARCHAR(255),
    role VARCHAR(255),
    custom_variables JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS emails (
    id BIGSERIAL PRIMARY KEY,
    campaign_id BIGINT REFERENCES campaigns(id) ON DELETE CASCADE,
    recipient_id BIGINT REFERENCES campaign_recipients(id) ON DELETE CASCADE,
    gmail_message_id VARCHAR(255),
    gmail_thread_id VARCHAR(255),
    subject TEXT,
    body TEXT,
    status VARCHAR(50) DEFAULT 'QUEUED',
    error_message TEXT,
    sent_at TIMESTAMP,
    opened_at TIMESTAMP,
    clicked_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
