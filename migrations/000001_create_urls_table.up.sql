CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS urls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    short_code VARCHAR(20) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    clicks BIGINT DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT short_code_length CHECK (LENGTH(short_code) >= 3)
);

-- Create indexes separately
CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);

CREATE INDEX IF NOT EXISTS idx_created_at ON urls(created_at);

CREATE INDEX IF NOT EXISTS idx_expires_at ON urls(expires_at);

-- Fixed function syntax (no space between $$)
CREATE
OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = NOW();

RETURN NEW;

END;

$$ language 'plpgsql';

CREATE TRIGGER update_urls_updated_at BEFORE
UPDATE
    ON urls FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();