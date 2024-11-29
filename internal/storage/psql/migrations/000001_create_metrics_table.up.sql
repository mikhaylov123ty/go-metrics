BEGIN;

CREATE TABLE IF NOT EXISTS metrics(
    unique_id VARCHAR(128) PRIMARY KEY NOT NULL,
    type VARCHAR(10) NOT NULL,
    name TEXT NOT NULL,
    value DOUBLE PRECISION,
    delta INTEGER
);

CREATE INDEX metrics_name ON metrics (name);
CREATE INDEX metrics_type ON metrics (type);

COMMIT ;