package cmd

import (
	"cli_backup_tool/internal/logging"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	sourcePath string
	destPath   []string
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backs up files from source to destination",
	Long: `Backup command copies files and folders from a source directory 
	to one or more destination directory. Supports full backup now.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate input
		if sourcePath == "" || len(destPath) == 0 {
			logging.ErrorLogger.Println("--source and --dest are required.")
			os.Exit(1)
		}

		// Convert to absolute paths
		absSource, err := filepath.Abs(sourcePath)
		if err != nil {
			logging.ErrorLogger.Printf("Error resolving source path: %v\n", err)
			os.Exit(1)
		}

		logging.InfoLogger.Printf("Backing up from %s\n", absSource)

		for i := 0; i < len(destPath); i++ {
			absDest, err := filepath.Abs(destPath[i])
			if err != nil {
				logging.ErrorLogger.Printf("Error resolving destination path: %v\n", err)
				os.Exit(1)
			}
			logging.InfoLogger.Printf("-> to: %s\n", absDest)
		}

		logging.InfoLogger.Println("Backup logic not implemented yet.")
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)

	backupCmd.Flags().StringVarP(&sourcePath, "source", "s", "s", "Source directory to back up (required)")
	backupCmd.Flags().StringSliceVarP(&destPath, "dest", "d", []string{}, "Destination directories for the backup (comma-separated)")
}
