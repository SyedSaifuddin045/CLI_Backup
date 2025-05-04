package backupengine

import (
	"os"
	"sync"

	"cli_backup_tool/internal/logging"
)

type CopyBackupStrategy struct{}

func NewCopyBackupStrategy() *CopyBackupStrategy {
	return &CopyBackupStrategy{}
}

// Backup performs a recursive copy from source to each destination
func (r *CopyBackupStrategy) Backup(source string, destinations []string) error {
	// No need for fmt package import - using only the logger

	// Create a wait group to wait for all backup operations to complete
	var wg sync.WaitGroup

	// Create an error channel to collect errors from goroutines
	errChan := make(chan error, len(destinations))

	// Launch a goroutine for each destination
	for _, dest := range destinations {
		wg.Add(1)

		// Capture the destination variable to avoid data race
		destination := dest

		go func() {
			defer wg.Done()

			logging.InfoLogger.Printf("Starting concurrent backup to %s\n", destination)

			// Define the file handling function for this destination
			fileProcessor := func(sourcePath, destPath string, info os.FileInfo) error {
				err := copyFile(sourcePath, destPath, info)
				if err != nil {
					logging.ErrorLogger.Printf("Error copying file %s: %v\n", sourcePath, err)
				}
				return err
			}

			// Define the directory handling function for this destination
			dirProcessor := func(sourcePath, destPath string, info os.FileInfo) error {
				err := os.MkdirAll(destPath, info.Mode())
				if err != nil {
					logging.ErrorLogger.Printf("Error creating directory %s: %v\n", destPath, err)
				}
				return err
			}

			// Use the utility function to walk the directory
			err := WalkSourceToDest(source, destination, fileProcessor, dirProcessor)

			if err != nil {
				logging.ErrorLogger.Printf("Failed to back up to %s: %v\n", destination, err)
				errChan <- err
				return
			}

			logging.InfoLogger.Printf("Backup to %s completed successfully\n", destination)
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Close the error channel
	close(errChan)

	// Collect all errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	// Return the first error if any occurred
	if len(errs) > 0 {
		logging.ErrorLogger.Printf("Errors occurred during concurrent backup: %v\n", errs)
		return errs[0]
	}

	return nil
}
