package migrate

import (
	"fmt"

	"github.com/thieman/steam-automigrate/internal/config"
)

func DoMigrate() error {
	config, err := config.GetConfig()
	if err != nil {
		return err
	}

	migrations, err := getMigrations(config)
	if err != nil {
		return err
	}

	fmt.Println(migrations)

	return nil
}
