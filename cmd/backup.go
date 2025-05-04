package cmd

import (
	"cli_backup_tool/internal/backupengine"
	"cli_backup_tool/internal/logging"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	sourcePath string
	destPath   []string
	strategy   string
	useCloud   bool
	useFTP     bool
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backs up files from source to destination",
	Long: `Backup command copies files and folders from a source directory 
	to one or more destination directories. Supports multiple strategies.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate input
		if sourcePath == "" || len(destPath) == 0 {
			logging.ErrorLogger.Println("--source and --dest are required.")
			os.Exit(1)
		}

		// Convert source to absolute path
		absSource, err := filepath.Abs(sourcePath)
		if err != nil {
			logging.ErrorLogger.Printf("Error resolving source path: %v\n", err)
			os.Exit(1)
		}

		// Convert destinations to absolute paths
		var absDests []string
		for _, d := range destPath {
			absDest, err := filepath.Abs(d)
			if err != nil {
				logging.ErrorLogger.Printf("Error resolving destination path: %v\n", err)
				os.Exit(1)
			}
			absDests = append(absDests, absDest)
		}

		// Select strategy
		var backup backupengine.BackupStrategy
		switch strategy {
		case "copy":
			backup = backupengine.NewCopyBackupStrategy()
		case "compressed":
			logging.ErrorLogger.Println("Compressed backup strategy not yet implemented.")
			os.Exit(1)
		default:
			logging.ErrorLogger.Printf("Unknown backup strategy: %s\n", strategy)
			os.Exit(1)
		}

		// Execute backup
		if err := backup.Backup(absSource, absDests); err != nil {
			logging.ErrorLogger.Printf("Backup failed: %v\n", err)
			os.Exit(1)
		}

		logging.InfoLogger.Println("Backup completed successfully.")
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)

	backupCmd.Flags().StringVarP(&sourcePath, "source", "s", "", "Source directory to back up (required)")
	backupCmd.Flags().StringSliceVarP(&destPath, "dest", "d", []string{}, "Destination directories (comma-separated)")

	backupCmd.Flags().StringVarP(&strategy, "strategy", "t", "copy", "Backup strategy: copy, compressed (default: copy)")
	backupCmd.Flags().BoolVar(&useCloud, "cloud", false, "Enable cloud backup (not yet implemented)")
	backupCmd.Flags().BoolVar(&useFTP, "ftp", false, "Enable FTP backup (not yet implemented)")
}
