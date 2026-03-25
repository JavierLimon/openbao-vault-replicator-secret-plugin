package replicator

import "fmt"

var (
	Version   = "1.0.0"
	Commit    = ""
	Date      = ""
	BuildType = ""
)

func GetVersion() string {
	return Version
}

func GetCommit() string {
	return Commit
}

func GetDate() string {
	return Date
}

func GetBuildType() string {
	return BuildType
}

func GetVersionInfo() string {
	return fmt.Sprintf("Version: %s, Commit: %s, Date: %s, BuildType: %s",
		Version, Commit, Date, BuildType)
}
