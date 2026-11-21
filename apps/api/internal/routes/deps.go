package routes

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/suprimkhatri77/turgorepo/api/internal/background"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/cloudinary"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/email"
	apiredis "github.com/suprimkhatri77/turgorepo/api/internal/packages/redis"
)

// Deps are what the routes need from the app.
type Deps struct {
	Config  *config.Config
	Queries *db.Queries
	Pool    *pgxpool.Pool
	Redis   *apiredis.Client
	Mailer  email.Mailer
	Jobs    *background.Runner
	// Cloudinary isn't used by any route yet; it's wired in for the first upload
	// endpoint. It's nil when Cloudinary isn't configured.
	Cloudinary *cloudinary.Client
}
