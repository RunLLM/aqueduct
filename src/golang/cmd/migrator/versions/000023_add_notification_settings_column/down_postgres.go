package _000023_add_notification_settings_column

const downPostgresScript = `
ALTER TABLE workflow DROP COLUMN IF EXISTS notification_settings;
`
