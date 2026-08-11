package migrations

import "embed"

// UpMigrationsFS embeds all .up.sql database migration scripts
//go:embed *.up.sql
var UpMigrationsFS embed.FS

