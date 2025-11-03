-- Development database initialization script
-- Database: meshprobe

-- Create database schema
CREATE SCHEMA IF NOT EXISTS meshprobe;

-- Users table (for future authentication)
CREATE TABLE IF NOT EXISTS meshprobe.users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Configurations table
CREATE TABLE IF NOT EXISTS meshprobe.configurations (
    id SERIAL PRIMARY KEY,
    config_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    version INTEGER DEFAULT 1,
    description TEXT,
    config_data JSONB NOT NULL,
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Probes table
CREATE TABLE IF NOT EXISTS meshprobe.probes (
    id SERIAL PRIMARY KEY,
    probe_id VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    port INTEGER DEFAULT 0,
    enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 10,
    timeout VARCHAR(20) DEFAULT '5s',
    interval_seconds INTEGER DEFAULT 1,
    expected_rtt VARCHAR(20),
    min_rtt VARCHAR(20),
    max_rtt VARCHAR(20),
    max_loss_pct DECIMAL(5,2) DEFAULT 1.0,
    tags TEXT[],
    region VARCHAR(100),
    environment VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Measurements table
CREATE TABLE IF NOT EXISTS meshprobe.measurements (
    id SERIAL PRIMARY KEY,
    probe_id VARCHAR(100) NOT NULL REFERENCES meshprobe.probes(probe_id),
    measurement_type VARCHAR(50) DEFAULT 'icmp',
    success BOOLEAN NOT NULL,
    response_time_ms DECIMAL(10,3),
    packet_loss_pct DECIMAL(5,2),
    ttl INTEGER,
    sequence_number INTEGER,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB
);

-- Sessions table (for JWT/session management)
CREATE TABLE IF NOT EXISTS meshprobe.sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES meshprobe.users(id),
    session_token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Health checks table
CREATE TABLE IF NOT EXISTS meshprobe.health_checks (
    id SERIAL PRIMARY KEY,
    component VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    response_time_ms DECIMAL(10,3),
    error_message TEXT,
    metadata JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert development data
INSERT INTO meshprobe.users (username, email, password_hash, role) VALUES
    ('admin', 'admin@localhost', '$2a$10$encrypted_password_hash', 'admin'),
    ('developer', 'dev@localhost', '$2a$10$encrypted_password_hash', 'developer')
ON CONFLICT (username) DO NOTHING;

-- Insert default configuration
INSERT INTO meshprobe.configurations (config_id, name, version, config_data, is_active) VALUES
    ('dev-default', 'Development Default Config', 1, 
     '{"targets": [], "network": {}, "telemetry": {"log_level": "debug"}}'::jsonb, true)
ON CONFLICT (config_id) DO NOTHING;

-- Insert sample probes
INSERT INTO meshprobe.probes (probe_id, display_name, address, port, enabled, priority, timeout, interval_seconds, tags, region, environment) VALUES
    ('google_dns_dev', 'Google DNS (Dev)', '8.8.8.8', 0, true, 10, '5s', 1, ARRAY['dns', 'public'], 'us-west', 'development'),
    ('cloudflare_dns_dev', 'Cloudflare DNS (Dev)', '1.1.1.1', 0, true, 8, '5s', 1, ARRAY['dns', 'public'], 'global', 'development'),
    ('localhost_dev', 'Localhost (Dev)', '127.0.0.1', 0, true, 5, '3s', 5, ARRAY['local', 'test'], 'local', 'development')
ON CONFLICT (probe_id) DO NOTHING;

-- Insert sample measurements
INSERT INTO meshprobe.measurements (probe_id, measurement_type, success, response_time_ms, packet_loss_pct, ttl, sequence_number, metadata) VALUES
    ('google_dns_dev', 'icmp', true, 12.5, 0.0, 64, 1, '{"size": 64}'),
    ('cloudflare_dns_dev', 'icmp', true, 15.2, 0.0, 64, 2, '{"size": 64}'),
    ('localhost_dev', 'icmp', true, 0.1, 0.0, 128, 3, '{"size": 64}')
ON CONFLICT DO NOTHING;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_measurements_probe_id ON meshprobe.measurements(probe_id);
CREATE INDEX IF NOT EXISTS idx_measurements_timestamp ON meshprobe.measurements(timestamp);
CREATE INDEX IF NOT EXISTS idx_measurements_probe_timestamp ON meshprobe.measurements(probe_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_health_checks_component ON meshprobe.health_checks(component);
CREATE INDEX IF NOT EXISTS idx_health_checks_timestamp ON meshprobe.health_checks(timestamp);

-- Create functions for automatic timestamp updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON meshprobe.users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_configurations_updated_at BEFORE UPDATE ON meshprobe.configurations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_probes_updated_at BEFORE UPDATE ON meshprobe.probes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Success message
SELECT 'Development database initialized successfully!' as message;