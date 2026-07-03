package app

import (
	"github.com/suprimkhatri77/turgorepo/api/internal/cron"
	dbgen "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
)

func initCron(queries *dbgen.Queries) {
	cron.CronExample(queries)
}
