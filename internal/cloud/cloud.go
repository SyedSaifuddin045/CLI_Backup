package cloud

import (
	"cli_backup_tool/internal/common"
	"cli_backup_tool/internal/logging"
)

func HandleCloudZipBackup(source string, destination common.DestinationStruct, zipFileName string) error {
	var err error
	logging.InfoLogger.Printf("Starting ZIP backup to cloud platform: %v, path: %s\n", destination.PlatformName, destination.Path)
	switch destination.PlatformName {
	case common.AWS:
		err = UploadToAWS()
	case common.AZURE:
		err = UploadToAzure()
	case common.GDRIVE:
		err = UploadToGDrive()
	default:
	}
	if err != nil {
		logging.DebugLogger.Printf("failed to upload to cloud storage: %s", err)
		return err
	}

	return nil
}

func UploadToGDrive() error {
	panic("unimplemented")
}

func UploadToAzure() error {
	panic("unimplemented")
}

func UploadToAWS() error {
	panic("unimplemented")
}
