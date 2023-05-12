package _000001_base

const sqliteScript = `
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER NOT NULL PRIMARY KEY,
    dirty BOOL NOT NULL,
    name TEXT NOT NULL
);

-- The schema_version record for v000001 needs to be explicitly inserted.
-- Only insert the record if it doesn't already exist.
INSERT OR IGNORE INTO schema_version VALUES (1, 1, 'base');

CREATE TABLE IF NOT EXISTS app_user (
    id BLOB NOT NULL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    organization_id TEXT NOT NULL,
    role TEXT NOT NULL,
    api_key TEXT NOT NULL UNIQUE,
    auth0_id TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS integration (
    id BLOB NOT NULL PRIMARY KEY,
    organization_id TEXT NOT NULL,
    service TEXT NOT NULL,
    name TEXT NOT NULL,
    config BLOB NOT NULL,
    created_at DATETIME NOT NULL,
    validated BOOL NOT NULL
);

CREATE TABLE IF NOT EXISTS notification (
    id BLOB NOT NULL PRIMARY KEY,
    receiver_id BLOB NOT NULL REFERENCES app_user (id),
    content TEXT NOT NULL,
    status TEXT NOT NULL,
    level TEXT NOT NULL,
    association BLOB NOT NULL,
    created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS workflow (
    id BLOB NOT NULL PRIMARY KEY,
    user_id BLOB NOT NULL REFERENCES app_user (id),
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    schedule BLOB NOT NULL,
    created_at DATETIME NOT NULL,
    UNIQUE (user_id, name)
);

CREATE TABLE IF NOT EXISTS workflow_dag (
    id BLOB NOT NULL PRIMARY KEY,
    workflow_id BLOB NOT NULL REFERENCES workflow (id),
    s3_config BLOB NOT NULL,
    created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS operator (
    id BLOB NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    spec BLOB NOT NULL
);

CREATE TABLE IF NOT EXISTS artifact (
    id BLOB NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    spec BLOB NOT NULL
);

CREATE TABLE IF NOT EXISTS workflow_dag_edge (
    workflow_dag_id BLOB NOT NULL REFERENCES workflow_dag (id),
    type TEXT NOT NULL,
    from_id BLOB NOT NULL,
    to_id BLOB NOT NULL,
    idx INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS workflow_dag_result (
    id BLOB NOT NULL PRIMARY KEY,
    workflow_dag_id BLOB NOT NULL REFERENCES workflow_dag (id),
    status TEXT NOT NULL,
    created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS operator_result (
    id BLOB NOT NULL PRIMARY KEY,
    workflow_dag_result_id BLOB NOT NULL REFERENCES workflow_dag_result (id),
    operator_id BLOB NOT NULL REFERENCES operator (id),
    status TEXT NOT NULL,
    metadata BLOB
);

CREATE TABLE IF NOT EXISTS artifact_result (
    id BLOB NOT NULL PRIMARY KEY,
    workflow_dag_result_id BLOB NOT NULL REFERENCES workflow_dag_result (id),
    artifact_ids BLOB NOT NULL REFERENCES artifact (id),
    content_path TEXT NOT NULL,
    status TEXT NOT NULL,
    metadata BLOB
);

CREATE TABLE IF NOT EXISTS workflow_watcher (
    workflow_id BLOB NOT NULL REFERENCES workflow (id),
    user_id BLOB NOT NULL REFERENCES app_user (id),
    PRIMARY KEY (workflow_id, user_id)
);
`
