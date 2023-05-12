package _000001_base

const upPostgresScript = `
-- Necessary for UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS schema_version (
    version BIGINT NOT NULL PRIMARY KEY,
    dirty BOOLEAN NOT NULL,
    name VARCHAR NOT NULL
);

-- The schema_version record for v000001 needs to be explicitly inserted.
-- Only insert the record if it doesn't already exist.
INSERT INTO schema_version VALUES (1, true, 'base') ON CONFLICT (version) DO NOTHING;

CREATE TABLE IF NOT EXISTS app_user (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    email VARCHAR NOT NULL UNIQUE,
    organization_id VARCHAR NOT NULL,
    role VARCHAR NOT NULL,
    api_key VARCHAR NOT NULL UNIQUE,
    auth0_id VARCHAR NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS integration (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id VARCHAR NOT NULL,
    service VARCHAR NOT NULL,
    name VARCHAR NOT NULL,
    config JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    validated BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS notification (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    receiver_id UUID NOT NULL REFERENCES app_user (id),
    content VARCHAR NOT NULL,
    status VARCHAR NOT NULL,
    level VARCHAR NOT NULL,
    association JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS workflow (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES app_user (id),
    name VARCHAR NOT NULL,
    description VARCHAR NOT NULL,
    schedule JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (user_id, name)
);

CREATE TABLE IF NOT EXISTS workflow_dag (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workflow_id UUID NOT NULL REFERENCES workflow (id),
    s3_config JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS operator (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR NOT NULL,
    description VARCHAR NOT NULL,
    spec JSONB NOT NULL
);

CREATE TABLE IF NOT EXISTS artifact (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR NOT NULL,
    description VARCHAR NOT NULL,
    spec JSONB NOT NULL
);

CREATE TABLE IF NOT EXISTS workflow_dag_edge (
    workflow_dag_id UUID NOT NULL REFERENCES workflow_dag (id),
    type VARCHAR NOT NULL,
    from_id UUID NOT NULL,
    to_id UUID NOT NULL,
    idx SMALLINT NOT NULL
);

CREATE TABLE IF NOT EXISTS workflow_dag_result (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workflow_dag_id UUID NOT NULL REFERENCES workflow_dag (id),
    status VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS operator_result (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workflow_dag_result_id UUID NOT NULL REFERENCES workflow_dag_result (id),
    operator_id UUID NOT NULL REFERENCES operator (id),
    status VARCHAR NOT NULL,
    metadata JSONB
);

CREATE TABLE IF NOT EXISTS artifact_result (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workflow_dag_result_id UUID NOT NULL REFERENCES workflow_dag_result (id),
    artifact_ids UUID NOT NULL REFERENCES artifact (id),
    content_path VARCHAR NOT NULL,
    status VARCHAR NOT NULL,
    metadata JSONB
);

CREATE TABLE IF NOT EXISTS workflow_watcher (
    workflow_id UUID NOT NULL REFERENCES workflow (id),
    user_id UUID NOT NULL REFERENCES app_user (id),
    PRIMARY KEY (workflow_id, user_id)
);
`
