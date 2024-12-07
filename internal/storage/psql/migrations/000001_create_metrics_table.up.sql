BEGIN;

CREATE TABLE IF NOT EXISTS metrics(
    name TEXT PRIMARY KEY NOT NULL,
    type VARCHAR(10) NOT NULL,
    value DOUBLE PRECISION,
    delta BIGINT
);

CREATE INDEX metrics_type ON metrics (type);

COMMIT ;