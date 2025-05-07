package backupengine

import (
	"archive/zip"
	"os"
	"path/filepath"
	"sync"
	"time"

	// You'll need to create this package

	"cli_backup_tool/internal/cloud"
	"cli_backup_tool/internal/common"
	"cli_backup_tool/internal/logging"
)

// ZipBackupStrategy implements a backup strategy that creates zip archives
type ZipBackupStrategy struct {
	// Configuration options
	CompressionLevel int
	TimeFormat       string
	TempDir          string // Directory for temporary files when using cloud storage
}

// NewZipBackupStrategy creates a new ZIP backup strategy with default settings
func NewZipBackupStrategy() *ZipBackupStrategy {
	return &ZipBackupStrategy{
		CompressionLevel: 6,
		TimeFormat:       "15-04-05_02-01-2006", // Default time format for zip file names
		TempDir:          os.TempDir(),          // Use system temp directory by default
	}
}

// Backup creates zip archives of the source directory at each destination
func (z *ZipBackupStrategy) Backup(source string, destinations []common.DestinationStruct) error {
	// Create a wait group to wait for all backup operations to complete
	var wg sync.WaitGroup

	// Create an error channel to collect errors from goroutines
	errChan := make(chan error, len(destinations))

	// Get the source directory name for the zip file
	sourceDirName := filepath.Base(source)

	// Generate timestamp for the zip file name
	timestamp := time.Now().Format(z.TimeFormat)

	// Launch a goroutine for each destination
	for _, dest := range destinations {
		wg.Add(1)
		// Capture the destination variable to avoid data race
		destination := dest

		go func() {
			defer wg.Done()

			zipFileName := sourceDirName + "_" + timestamp + ".zip"

			if destination.IsCloud {
				// Handle cloud storage backup
				err := cloud.HandleCloudZipBackup(source, destination, zipFileName)
				if err != nil {
					logging.ErrorLogger.Printf("Failed cloud backup to %s: %v\n", destination.Path, err)
					errChan <- err
				} else {
					logging.InfoLogger.Printf("Successfully completed cloud backup to %v: %s\n",
						destination.PlatformName, destination.Path)
				}
			} else {
				// Handle local storage backup
				err := z.handleLocalBackup(source, destination, zipFileName)
				if err != nil {
					logging.ErrorLogger.Printf("Failed local backup to %s: %v\n", destination.Path, err)
					errChan <- err
				} else {
					logging.InfoLogger.Printf("Successfully completed local backup to %s\n", destination.Path)
				}
			}
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
		logging.ErrorLogger.Printf("Errors occurred during ZIP backup: %v\n", errs)
		return errs[0]
	}

	return nil
}

// handleLocalBackup manages backup to local storage
func (z *ZipBackupStrategy) handleLocalBackup(source string, destination common.DestinationStruct, zipFileName string) error {
	fullZipPath := filepath.Join(destination.Path, zipFileName)
	logging.InfoLogger.Printf("Starting ZIP backup to local path %s\n", fullZipPath)

	// Ensure destination directory exists
	err := os.MkdirAll(filepath.Dir(fullZipPath), 0755)
	if err != nil {
		logging.ErrorLogger.Printf("Failed to create destination directory: %v\n", err)
		return err
	}

	// Create a new zip file
	zipFile, err := os.Create(fullZipPath)
	if err != nil {
		logging.ErrorLogger.Printf("Failed to create zip file %s: %v\n", fullZipPath, err)
		return err
	}
	defer zipFile.Close()

	// Create a zip writer with the specified compression level
	zipWriter := zip.NewWriter(zipFile)

	// Process each file and add it to the zip archive
	err = CreateZipArchive(source, zipWriter)
	if err != nil {
		logging.ErrorLogger.Printf("Failed to create zip archive: %v\n", err)
		return err
	}

	// Close the zip writer to finalize the archive
	err = zipWriter.Close()
	if err != nil {
		logging.ErrorLogger.Printf("Failed to finalize zip archive: %v\n", err)
		return err
	}

	logging.InfoLogger.Printf("ZIP backup to %s completed successfully\n", fullZipPath)
	return nil
}
