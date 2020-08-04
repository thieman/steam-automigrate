package migrate

import (
	"sort"
	"strconv"
	"time"

	"github.com/thieman/steam-automigrate/internal/steam"
)

type Migration struct {
	AppID           string
	AppName         string
	InstallDir      string
	SizeOnDiskBytes int64
	FromLibrary     *steam.Library
	ToLibrary       *steam.Library
}

func getDesiredLocations(config *steam.Config) (map[string]string, error) {
	playtimes, err := steam.GetPlaytimes(config)
	if err != nil {
		return nil, err
	}

	desired := make(map[string]string)
	thresholdDuration, err := time.ParseDuration("-" + strconv.Itoa(config.MigrateThresholdSeconds) + "s")
	if err != nil {
		return nil, err
	}

	for appID, lastPlayed := range playtimes {
		desiredLocation := "ssd"
		if time.Now().Add(thresholdDuration).After(lastPlayed) {
			desiredLocation = "hdd"
		}

		desired[appID] = desiredLocation
	}

	return desired, nil
}

func getMigrations(config *steam.Config) ([]Migration, error) {
	desired, err := getDesiredLocations(config)
	if err != nil {
		return nil, err
	}

	installed, err := steam.GetInstalledApps(config)
	if err != nil {
		return nil, err
	}

	var migrations []Migration

	sort.Slice(installed, func(i, j int) bool { return installed[i].SizeOnDiskBytes > installed[j].SizeOnDiskBytes })

	for _, app := range installed {
		desiredType, found := desired[app.AppID]
		if !found || desiredType == app.Library.Type {
			continue
		}
	}

	return migrations, nil
}
