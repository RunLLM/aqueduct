package _00026_drop_integration_validated_column

const upPostgresScript = `
ALTER TABLE integration 
DROP COLUMN IF EXISTS validated;
`
