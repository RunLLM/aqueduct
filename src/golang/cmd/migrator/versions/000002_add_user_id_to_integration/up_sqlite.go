package _000002_add_user_id_to_integration

const sqliteScript = `
ALTER TABLE integration ADD COLUMN user_id BLOB REFERENCES app_user (id);
`
