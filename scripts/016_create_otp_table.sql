-- Create OTP codes table for email verification

CREATE TABLE IF NOT EXISTS yt_otp_codes (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    code VARCHAR(6) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    attempts INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_otp_email ON yt_otp_codes(email);
CREATE INDEX IF NOT EXISTS idx_otp_expires ON yt_otp_codes(expires_at);

-- Add email configuration table
CREATE TABLE IF NOT EXISTS yt_email_config (
    id SERIAL PRIMARY KEY,
    smtp_host VARCHAR(255) DEFAULT 'smtp.gmail.com',
    smtp_port INTEGER DEFAULT 587,
    smtp_username VARCHAR(255),
    smtp_password VARCHAR(255),
    from_email VARCHAR(255) DEFAULT 'noreply@yukti.io',
    from_name VARCHAR(255) DEFAULT 'Yukti FinOps',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Insert default email config
INSERT INTO yt_email_config (smtp_host, smtp_port, from_email, from_name)
VALUES ('smtp.gmail.com', 587, 'noreply@yukti.io', 'Yukti FinOps')
ON CONFLICT DO NOTHING;

SELECT 'OTP and email tables created successfully' as status;