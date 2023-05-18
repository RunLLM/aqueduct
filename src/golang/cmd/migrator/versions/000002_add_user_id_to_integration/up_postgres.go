package _000002_add_user_id_to_integration

const upPostgresScript = `
ALTER TABLE integration ADD COLUMN user_id UUID;

-- Add foreign key constraint with minimal blocking
ALTER TABLE integration ADD CONSTRAINT integration_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES app_user (id)
    NOT VALID;

ALTER TABLE integration VALIDATE CONSTRAINT integration_user_id_fkey;
`
