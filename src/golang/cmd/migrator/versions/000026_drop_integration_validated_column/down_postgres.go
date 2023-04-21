package _00026_drop_integration_validated_column

const downPostgresScript = `
ALTER TABLE integration ADD COLUMN validated BOOLEAN;
`
