package migrate

import (
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/thieman/steam-automigrate/internal/steam"
)

const LIBRARY_RESERVED_BYTES = 5 * 1024 * 1024 * 1024 // 5 GB

type Migration struct {
	App       *steam.InstalledApp
	ToLibrary *steam.Library
}

type migrationDeltas map[*steam.Library]int64

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

func getTargetLibrary(config *steam.Config, installed []*steam.InstalledApp, app *steam.InstalledApp, desiredType string, deltas migrationDeltas) *steam.Library {
	candidates := config.SSDs
	if desiredType == "hdd" {
		candidates = config.HDDs
	}

	// Current algo is to prefer to balance library size (NOT free space remaining) evenly
	// across all disks of a type. Might add options between different algos later on.
	var target *steam.Library
	var targetLibrarySize uint64

	for _, library := range candidates {
		librarySize := uint64(int64(steam.TotalSizeOfLibraryBytes(library, installed)) + deltas[library])
		committedSpace := librarySize + LIBRARY_RESERVED_BYTES
		if committedSpace+app.SizeOnDiskBytes > library.TotalSpaceBytes {
			continue
		}

		if target == nil || targetLibrarySize > librarySize {
			target = library
			targetLibrarySize = librarySize
		}

	}

	return target
}

func getMigrations(config *steam.Config) ([]*Migration, migrationDeltas, error) {
	if len(config.SSDs) == 0 || len(config.HDDs) == 0 {
		return nil, nil, errors.New("Cannot migrate unless at least one SSD and one HDD are defined")
	}

	desired, err := getDesiredLocations(config)
	if err != nil {
		return nil, nil, err
	}

	installed, err := steam.GetInstalledApps(config)
	if err != nil {
		return nil, nil, err
	}

	var migrations []*Migration

	sort.Slice(installed, func(i, j int) bool { return installed[i].SizeOnDiskBytes > installed[j].SizeOnDiskBytes })

	deltas := make(migrationDeltas)
	for _, library := range append(config.SSDs, config.HDDs...) {
		deltas[library] = 0
	}

	for _, app := range installed {
		desiredType, found := desired[app.AppID]
		if !found {
			// If we've installed a game but never played it, it won't be in the playtime data
			// and therefore won't show up in desired here. Default anything unplayed to HDD.
			desiredType = "hdd"
		}

		if desiredType == app.Library.Type {
			continue
		}

		toLibrary := getTargetLibrary(config, installed, app, desiredType, deltas)
		if toLibrary == nil { // no suitable target
			continue
		}

		migrations = append(migrations, &Migration{App: app, ToLibrary: toLibrary})
		deltas[toLibrary] += int64(app.SizeOnDiskBytes)
		deltas[app.Library] -= int64(app.SizeOnDiskBytes)
	}

	return migrations, deltas, nil
}
