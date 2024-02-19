CREATE SCHEMA server;

CREATE TABLE server.metrics (
    id SERIAL NOT NULL PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    type VARCHAR(256) NOT NULL,
    value DOUBLE PRECISION,
    delta BIGINT
);

ALTER TABLE server.metrics ADD CONSTRAINT unique_metrics_name_type UNIQUE (name, type);