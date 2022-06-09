package _000002_add_user_id_to_integration

const downPostgresScript = `
ALTER TABLE integration DROP CONSTRAINT IF EXISTS integration_user_id_fkey;

ALTER TABLE integration DROP COLUMN IF EXISTS user_id;
`
