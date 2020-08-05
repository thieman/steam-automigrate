package migrate

import (
	"fmt"
	"sort"
	"sync"

	"github.com/thieman/steam-automigrate/internal/steam"
)

const SAFETY_BUFFER_BYTES = 2 * 1024 * 1024 * 1024

type migrateResult struct {
	migration *Migration
	err       error
}

func migrateApp(migration *Migration, wg *sync.WaitGroup, outputChannel chan migrateResult) {
	defer wg.Done()
	fmt.Printf("Migrate %v from %v to %v\n", migration.App.AppName, migration.App.Library.VolumeName, migration.ToLibrary.VolumeName)
	outputChannel <- migrateResult{migration, nil}
}

func migrate(config *steam.Config, migrations []*Migration) error {
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].App.SizeOnDiskBytes > migrations[j].App.SizeOnDiskBytes })

	// We ideally want to be running one migration on each HDD at all times for maximum utilization.
	// We also need to be cautious that we don't temporarily exceed a drive's capacity by moving
	// something TO it before we've moved enough FROM it. So each time we have an available slot
	// for migration, we'll go through and find a migration to do keeping all that in mind.
	migrating := make(map[*steam.Library]bool)
	freeSpace := make(map[*steam.Library]int64)
	for _, library := range append(config.SSDs, config.HDDs...) {
		if library.Type == "hdd" {
			migrating[library] = false
		}
		freeSpace[library] = int64(library.FreeSpaceBytes)
	}

	var wg sync.WaitGroup
	outputChannel := make(chan migrateResult)
	var firstError error

outer:
	for {
		var result migrateResult
		select {
		case result = <-outputChannel:
		default:
		}

		if result != (migrateResult{}) {
			if result.err != nil {
				firstError = result.err
				fmt.Println("Something went wrong, waiting for already-started migrations to finish")
				break outer
			}

			freeSpace[result.migration.App.Library] += int64(result.migration.App.SizeOnDiskBytes)
			if result.migration.ToLibrary.Type == "hdd" {
				migrating[result.migration.ToLibrary] = false
			} else {
				migrating[result.migration.App.Library] = false
			}
		}

		for i, migration := range migrations {
			if migration == result.migration {
				migrations[i] = nil
				break
			}
		}

		anyRemaining := false
		for _, migration := range migrations {
			if migration == nil {
				continue
			}
			anyRemaining = true

			hddLibrary := migration.App.Library
			if migration.ToLibrary.Type == "hdd" {
				hddLibrary = migration.ToLibrary
			}

			isMigrating := migrating[hddLibrary]
			if isMigrating {
				continue
			}

			// Need to keep track of total committed space. We add commit before a migration to the TO
			// drive, and only remove it from the FROM drive *after* the migration is finished.
			if int64(migration.App.SizeOnDiskBytes)+SAFETY_BUFFER_BYTES < freeSpace[migration.ToLibrary] {
				freeSpace[migration.ToLibrary] -= int64(migration.App.SizeOnDiskBytes)
				migrating[hddLibrary] = true
				wg.Add(1)
				go migrateApp(migration, &wg, outputChannel)
				continue outer
			}
		}

		if !anyRemaining {
			break
		}

		// If we got here, we have more migrations to do but we can't do them yet because we don't
		// have sufficient space. We'll temporarily block here on the output channel to wait, then
		// throw the result back on the redirectChannel so we don't have to rework this messy function more.
		result = <-outputChannel
		go func() { outputChannel <- result }()
	}

	wg.Wait()

	return firstError
}
