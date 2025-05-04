/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"cli_backup_tool/cmd"
	"cli_backup_tool/internal/logging"
)

func main() {
	logging.Init(true, true, "logs/cli_backup.log")

	cmd.Execute()
}
