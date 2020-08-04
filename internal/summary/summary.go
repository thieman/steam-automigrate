package summary

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/thieman/steam-automigrate/internal/steam"
)

func printTypeSummary(libraryType string, summaries []*steam.Library) {
	var total, free uint64

	for _, summary := range summaries {
		total += summary.TotalSpaceBytes
		free += summary.FreeSpaceBytes
	}

	gbFree := float64(free) / (1024 * 1024 * 1024)
	percentFree := 100 * float64(free) / float64(total)

	bold := color.New(color.Bold)
	bold.Printf(strings.ToUpper(libraryType)+"s: %2.0f%% Free (%2.0f GB)\n\n", percentFree, gbFree)
}

func printLibrarySummary(library *steam.Library) {
	gbFree := float64(library.FreeSpaceBytes) / (1024 * 1024 * 1024)
	fmt.Printf("    "+library.Path+": %2.0f%% Free (%2.0f GB)\n", library.PercentFree, gbFree)
}

func printDetails(library *steam.Library, installed []*steam.InstalledApp, detailed bool) {
	filtered := steam.AppsBySizeForLibrary(library, installed)

	var shown uint

	for _, app := range filtered {
		if !detailed && shown >= 5 {
			break
		}

		gb := float64(app.SizeOnDiskBytes) / (1024 * 1024 * 1024)
		fmt.Printf("        "+app.AppName+" (%2.1f GB)\n", gb)

		shown++
	}

	fmt.Println()
}

func DoSummary(detailed bool) error {
	config, err := steam.GetConfig()
	if err != nil {
		return err
	}

	installed, err := steam.GetInstalledApps(config)
	if err != nil {
		return err
	}

	for _, coll := range [][]*steam.Library{config.SSDs, config.HDDs} {
		if len(coll) == 0 {
			continue
		}

		printTypeSummary(coll[0].Type, coll)

		for _, library := range coll {
			printLibrarySummary(library)
			printDetails(library, installed, detailed)
		}
	}

	return nil
}
