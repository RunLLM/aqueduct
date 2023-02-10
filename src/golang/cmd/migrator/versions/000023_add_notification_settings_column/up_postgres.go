package _000023_add_notification_settings_column

const upPostgresScript = `
ALTER TABLE workflow 
ADD COLUMN notification_settings JSONB;
`
