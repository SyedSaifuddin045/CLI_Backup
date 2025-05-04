package backupengine

type BackupStrategy interface {
	Backup(source string, destinations []string) error
}
