package backupengine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cli_backup_tool/internal/logging"
)

// copyFile copies an individual file from src to dst with proper logging
func copyFile(src, dst string, info os.FileInfo) error {
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
