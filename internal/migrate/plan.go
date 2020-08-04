package migrate

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/andygrunwald/vdf"
	"github.com/thieman/steam-automigrate/internal/config"
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

type installedApp struct {
	AppID           string
	AppName         string
	InstallDir      string
	SizeOnDiskBytes int64
	Library         *steam.Library
}

func getDesiredLocations(config *config.Config) (map[string]string, error) {
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

func getInstalledApps(config *config.Config) ([]installedApp, error) {
	var installed []installedApp

	for _, libraryConfig := range append(config.HDDs, config.SSDs...) {
		library, err := steam.LibraryFromConfig(&libraryConfig)
		if err != nil {
			return nil, err
		}

		for _, manifestPath := range library.ManifestPaths {
			f, err := os.Open(manifestPath)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, err
			}

			p := vdf.NewParser(f)
			data, err := p.Parse()
			if err != nil {
				return nil, err
			}

			state := data["AppState"].(map[string]interface{})
			size, err := strconv.ParseInt(state["SizeOnDisk"].(string), 10, 64)
			if err != nil {
				return nil, err
			}

			app := installedApp{
				AppID:           state["appid"].(string),
				AppName:         state["name"].(string),
				InstallDir:      state["installdir"].(string),
				SizeOnDiskBytes: size,
				Library:         library,
			}

			installed = append(installed, app)
		}
	}

	return installed, nil
}

func getMigrations(config *config.Config) ([]Migration, error) {
	_, err := getDesiredLocations(config)
	if err != nil {
		return nil, err
	}

	installed, err := getInstalledApps(config)
	if err != nil {
		return nil, err
	}

	fmt.Println(installed)

	var migrations []Migration

	return migrations, nil
}
