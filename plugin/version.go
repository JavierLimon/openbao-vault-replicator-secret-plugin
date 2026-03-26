package replicator

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	Version   = "1.0.0"
	Commit    = ""
	Date      = ""
	BuildType = ""
	Branch    = ""
)

type VersionInfo struct {
	Version   string
	Commit    string
	Date      string
	BuildType string
	Branch    string
	GoVersion string
	Compiler  string
	Platform  string
}

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

func GetBranch() string {
	return Branch
}

func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		BuildType: BuildType,
		Branch:    Branch,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

func GetVersionString() string {
	v := GetVersionInfo()
	return fmt.Sprintf("Version: %s, Commit: %s, Date: %s, BuildType: %s, Branch: %s, Go: %s, Platform: %s",
		v.Version, v.Commit, v.Date, v.BuildType, v.Branch, v.GoVersion, v.Platform)
}

func GetShortVersionString() string {
	version := Version
	if Commit != "" {
		version = fmt.Sprintf("%s (%s)", version, shortCommit(Commit))
	}
	return version
}

func shortCommit(commit string) string {
	if len(commit) >= 7 {
		return commit[:7]
	}
	return commit
}

func IsVersionAtLeast(required string) bool {
	return compareVersions(Version, required) >= 0
}

func compareVersions(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	for i := 0; i < len(partsA) || i < len(partsB); i++ {
		partA := 0
		partB := 0
		if i < len(partsA) {
			fmt.Sscanf(partsA[i], "%d", &partA)
		}
		if i < len(partsB) {
			fmt.Sscanf(partsB[i], "%d", &partB)
		}
		if partA > partB {
			return 1
		}
		if partA < partB {
			return -1
		}
	}
	return 0
}
