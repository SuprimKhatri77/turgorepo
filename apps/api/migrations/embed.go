// Package migrations embeds the SQL migrations, so production images can run them
// with no extra files or tools.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
