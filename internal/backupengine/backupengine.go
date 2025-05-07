package backupengine

import "cli_backup_tool/internal/common"

type BackupStrategy interface {
	Backup(source string, destinations []common.DestinationStruct) error
}
