package migrate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/thieman/steam-automigrate/internal/steam"
)

func printTypeSummary(fromLibraryType string, migrations []*Migration) {
	var size, appsCount uint64
	for _, migration := range migrations {
		if migration.App.Library.Type != fromLibraryType {
			continue
		}
		appsCount++
		size += migration.App.SizeOnDiskBytes
	}

	gb := size / (1024 * 1024 * 1024)
	other := "HDD"
	if fromLibraryType == "hdd" {
		other = "SSD"
	}

	bold := color.New(color.Bold)
	bold.Printf("%vs: Will move %d apps totaling %.1d GB to %v\n\n", strings.ToUpper(fromLibraryType), appsCount, gb, other)
}

func printMigrationDetails(fromLibraryType string, migrations []*Migration) {
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].App.SizeOnDiskBytes > migrations[j].App.SizeOnDiskBytes })

	for _, migration := range migrations {
		if migration.App.Library.Type != fromLibraryType {
			continue
		}
		fmt.Printf("    %v (%.1fGB)\n", migration.App.AppName, float64(migration.App.SizeOnDiskBytes)/(1024*1024*1024))
	}
	fmt.Println()
}

func printPostMigrationSummary(config *steam.Config, apps []*steam.InstalledApp, migrations []*Migration, deltas migrationDeltas) {
	bold := color.New(color.Bold)
	bold.Printf("As a result of this migration, library sizes will change:\n\n")

	for _, library := range append(config.SSDs, config.HDDs...) {
		size := steam.TotalSizeOfLibraryBytes(library, apps)
		fmt.Printf("    %v: %.1d GB -> %.1d GB\n\n", library.Path, size/(1024*1024*1024), (int64(size)+deltas[library])/(1024*1024*1024))
	}
}

func DoPlan() error {
	config, err := steam.GetConfig()
	if err != nil {
		return err
	}

	apps, err := steam.GetInstalledApps(config)
	if err != nil {
		return err
	}

	migrations, deltas, err := getMigrations(config)
	if err != nil {
		return err
	}

	for _, libraryType := range []string{"ssd", "hdd"} {
		printTypeSummary(libraryType, migrations)
		printMigrationDetails(libraryType, migrations)
	}

	printPostMigrationSummary(config, apps, migrations, deltas)

	return nil
}
