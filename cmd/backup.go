package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	sourcePath string
	destPath   []string
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backs up files from source to destination",
	Long: `Backup command copies files and folders from a source directory 
	to one or more destination directory. Supports full backup now.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate input
		if sourcePath == "" || len(destPath) == 0 {
			fmt.Println("Error: --source and --dest are required.")
			os.Exit(1)
		}

		// Convert to absolute paths
		absSource, err := filepath.Abs(sourcePath)
		if err != nil {
			fmt.Printf("Error resolving source path: %v\n", err)
			os.Exit(1)
		}

		// Debug print for now
		fmt.Printf("Backing up from %s \n", absSource)

		for i := 0; i < len(destPath); i++ {
			absDest, err := filepath.Abs(destPath[i])
			if err != nil {
				fmt.Printf("Error resolving destination path: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("-> to :%s\n", absDest)
		}

		// TODO: Add actual backup logic
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)

	// Define flags for the backup command
	backupCmd.Flags().StringVarP(&sourcePath, "source", "s", "s", "Source directory to back up (required)")
	backupCmd.Flags().StringSliceVarP(&destPath, "dest", "d", []string{}, "Destination directories for the backup (comma-separated)")
}
