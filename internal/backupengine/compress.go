package backupengine

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"cli_backup_tool/internal/logging"
)

// ZipBackupStrategy implements a backup strategy that creates zip archives
type ZipBackupStrategy struct {
	// Configuration options could be added here
	CompressionLevel int
	TimeFormat       string
}

// NewZipBackupStrategy creates a new ZIP backup strategy with default settings
func NewZipBackupStrategy() *ZipBackupStrategy {
	return &ZipBackupStrategy{
		CompressionLevel: 6,
		TimeFormat:       "15-04-05_02-01-2006", // Default time format for zip file names
	}
}

// Backup creates zip archives of the source directory at each destination
func (z *ZipBackupStrategy) Backup(source string, destinations []string) error {
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

			// Create zip file name with timestamp
			zipFileName := filepath.Join(destination, sourceDirName+"_"+timestamp+".zip")
			logging.InfoLogger.Printf("Starting ZIP backup to %s\n", zipFileName)

			// Ensure destination directory exists
			err := os.MkdirAll(filepath.Dir(zipFileName), 0755)
			if err != nil {
				logging.ErrorLogger.Printf("Failed to create destination directory: %v\n", err)
				errChan <- err
				return
			}

			// Create a new zip file
			zipFile, err := os.Create(zipFileName)
			if err != nil {
				logging.ErrorLogger.Printf("Failed to create zip file %s: %v\n", zipFileName, err)
				errChan <- err
				return
			}
			defer zipFile.Close()

			// Create a zip writer with the specified compression level
			zipWriter := zip.NewWriter(zipFile)
			defer zipWriter.Close()

			// Process each file and add it to the zip archive
			err = z.createZipArchive(source, zipWriter)
			if err != nil {
				logging.ErrorLogger.Printf("Failed to create zip archive: %v\n", err)
				errChan <- err
				return
			}

			// Close the zip writer to finalize the archive
			err = zipWriter.Close()
			if err != nil {
				logging.ErrorLogger.Printf("Failed to finalize zip archive: %v\n", err)
				errChan <- err
				return
			}

			logging.InfoLogger.Printf("ZIP backup to %s completed successfully\n", zipFileName)
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

// createZipArchive walks through the source directory and adds files to the zip archive
func (z *ZipBackupStrategy) createZipArchive(source string, zipWriter *zip.Writer) error {
	// Define handlers for our utility function
	fileProcessor := func(sourcePath, destPath string, info os.FileInfo) error {
		// Get the relative path for the zip entry
		relPath, err := filepath.Rel(source, sourcePath)
		if err != nil {
			logging.ErrorLogger.Printf("Error getting relative path: %v\n", err)
			return err
		}

		// Use forward slashes in zip files regardless of OS
		relPath = filepath.ToSlash(relPath)

		// Create a zip file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			logging.ErrorLogger.Printf("Error creating zip header for %s: %v\n", sourcePath, err)
			return err
		}

		// Set the name with the relative path
		header.Name = relPath

		// Set compression method
		header.Method = zip.Deflate

		// Create the file entry in the zip
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			logging.ErrorLogger.Printf("Error creating zip entry for %s: %v\n", sourcePath, err)
			return err
		}

		// Open the source file
		file, err := os.Open(sourcePath)
		if err != nil {
			logging.ErrorLogger.Printf("Error opening file %s: %v\n", sourcePath, err)
			return err
		}
		defer file.Close()

		// Copy the file content to the zip entry
		_, err = io.Copy(writer, file)
		if err != nil {
			logging.ErrorLogger.Printf("Error copying file content to zip: %v\n", err)
			return err
		}

		return nil
	}

	// Directory processor - for zip files we don't need to create directories explicitly
	// as they are created implicitly when files are added
	dirProcessor := func(sourcePath, destPath string, info os.FileInfo) error {
		// For empty directories, we need to add them explicitly
		if isEmpty, err := isEmptyDir(sourcePath); err != nil {
			return err
		} else if isEmpty {
			// Get the relative path for the zip entry
			relPath, err := filepath.Rel(source, sourcePath)
			if err != nil {
				logging.ErrorLogger.Printf("Error getting relative path: %v\n", err)
				return err
			}

			// Use forward slashes in zip files regardless of OS
			relPath = filepath.ToSlash(relPath)

			// Ensure directory paths end with a slash
			if !strings.HasSuffix(relPath, "/") {
				relPath += "/"
			}

			// Create directory entry in zip
			_, err = zipWriter.Create(relPath)
			if err != nil {
				logging.ErrorLogger.Printf("Error creating directory entry in zip: %v\n", err)
				return err
			}
		}
		return nil
	}

	// Use our utility function to walk the directory
	return WalkSourceToDest(source, "", fileProcessor, dirProcessor)
}

// isEmptyDir checks if a directory is empty
func isEmptyDir(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// Read one entry from the directory
	_, err = f.Readdir(1)

	// If we got EOF, the directory is empty
	if err == io.EOF {
		return true, nil
	}

	// Any other error is propagated
	if err != nil {
		return false, err
	}

	// We read at least one entry, so the directory is not empty
	return false, nil
}
