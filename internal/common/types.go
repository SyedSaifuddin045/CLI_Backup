package common

type CloudPlatform int

const (
	UNKNOWN CloudPlatform = iota
	LOCAL
	AWS
	AZURE
	GDRIVE
)

type DestinationStruct struct {
	Path         string
	IsCloud      bool
	PlatformName CloudPlatform
}
