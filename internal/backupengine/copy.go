package backupengine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cli_backup_tool/internal/logging"
)

type CopyBackupStrategy struct{}

func NewCopyBackupStrategy() *CopyBackupStrategy {
	return &CopyBackupStrategy{}
}

// Backup performs a recursive copy from source to each destination
func (r *CopyBackupStrategy) Backup(source string, destinations []string) error {
	for _, dest := range destinations {
		logging.InfoLogger.Printf("Starting recursive backup to %s\n", dest)
		err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
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
				return os.MkdirAll(destPath, info.Mode())
			}

			return copyFile(path, destPath, info)
		})

		if err != nil {
			logging.ErrorLogger.Printf("Failed to back up to %s: %v\n", dest, err)
			return err
		}

		logging.InfoLogger.Printf("Backup to %s completed successfully\n", dest)
	}
	return nil
}

// copyFile copies an individual file from src to dst
func copyFile(src, dst string, info os.FileInfo) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer srcFile.Close()

	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directories for %s: %w", dst, err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file from %s to %s: %w", src, dst, err)
	}

	return nil
}
