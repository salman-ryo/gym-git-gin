package migrations

import _ "embed"

// MigrationSQL embeds the initial database creation DDL script
//go:embed 000001_create_tables.up.sql
var MigrationSQL string
