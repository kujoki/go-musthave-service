CREATE TABLE IF NOT EXISTS url_data (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    short_url VARCHAR(128) NOT NULL UNIQUE,
    long_url VARCHAR(1024) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_data_short_url ON url_data(short_url);

CREATE INDEX IF NOT EXISTS idx_data_long_url ON url_data(long_url);
