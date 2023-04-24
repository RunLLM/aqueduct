package _00026_drop_integration_validated_column

const sqliteScript = `
ALTER TABLE integration 
DROP COLUMN validated;
`
