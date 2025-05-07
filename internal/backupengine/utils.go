package backupengine

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cli_backup_tool/internal/logging"
)

// CopyFile copies an individual file from src to dst with proper logging
func CopyFile(src, dst string, info os.FileInfo) error {
	logging.DebugLogger.Printf("Copying file from %s to %s\n", src, dst)

	srcFile, err := os.Open(src)
	if err != nil {
		logging.ErrorLogger.Printf("Failed to open source file %s: %v\n", src, err)
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer srcFile.Close()

	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		logging.ErrorLogger.Printf("Failed to create directories for %s: %v\n", dst, err)
		return fmt.Errorf("failed to create directories for %s: %w", dst, err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		logging.ErrorLogger.Printf("Failed to create destination file %s: %v\n", dst, err)
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		logging.ErrorLogger.Printf("Failed to copy file from %s to %s: %v\n", src, dst, err)
		return fmt.Errorf("failed to copy file from %s to %s: %w", src, dst, err)
	}

	logging.DebugLogger.Printf("Successfully copied file to %s\n", dst)
	return nil
}

// FileProcessFunc defines a function signature for processing a file during backup
type FileProcessFunc func(sourcePath, destPath string, info os.FileInfo) error

// DirectoryProcessFunc defines a function signature for processing a directory during backup
type DirectoryProcessFunc func(sourcePath, destPath string, info os.FileInfo) error

// WalkSourceToDest walks through source directory and processes files and directories
// applying the provided functions for each type of item encountered
func WalkSourceToDest(source, dest string, fileFunc FileProcessFunc, dirFunc DirectoryProcessFunc) error {
	return filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				logging.ErrorLogger.Printf("Error accessing path %s: %v\n", path, err)
				return err
			}

			relPath, err := filepath.Rel(source, path)
			if err != nil {
				logging.ErrorLogger.Printf("Error getting relative path: %v\n", err)
				return err
			}

			destPath := filepath.Join(dest, relPath)

			if info.IsDir() {
				if dirFunc != nil {
					return dirFunc(path, destPath, info)
				}
				// Default directory behavior if no function is provided
				return os.MkdirAll(destPath, info.Mode())
			}

			if fileFunc != nil {
				return fileFunc(path, destPath, info)
			}

			// If no file function is provided, just return nil (do nothing)
			return nil
		})
}

// CreateZipArchive walks through the source directory and adds files to the zip archive
func CreateZipArchive(source string, zipWriter *zip.Writer) error {
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
