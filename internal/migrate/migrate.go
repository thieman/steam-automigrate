package migrate

import "github.com/thieman/steam-automigrate/internal/steam"

func DoMigrate() error {
	config, err := steam.GetConfig()
	if err != nil {
		return err
	}

	migrations, _, err := getMigrations(config)
	if err != nil {
		return err
	}

	return migrate(config, migrations)
}
