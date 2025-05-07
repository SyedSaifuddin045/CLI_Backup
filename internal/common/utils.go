package common

import "strings"

func PrepareDestinationStruct(dest string) DestinationStruct {
	switch {
	case strings.HasPrefix(dest, "s3://"):
		return DestinationStruct{Path: dest, IsCloud: true, PlatformName: AWS}
	case strings.HasPrefix(dest, "azure://"):
		return DestinationStruct{Path: dest, IsCloud: true, PlatformName: AZURE}
	case strings.HasPrefix(dest, "gdrive://"):
		return DestinationStruct{Path: dest, IsCloud: true, PlatformName: GDRIVE}
	default:
		return DestinationStruct{Path: dest, IsCloud: false, PlatformName: LOCAL}
	}
}
