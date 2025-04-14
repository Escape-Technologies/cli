package version

import "fmt"

var (
	Version   = "local"
	Commit    = "unknown"
	BuildDate = "unknown"
)

type VersionT struct {
	Version   string
	Commit    string
	BuildDate string
}

func GetVersion() VersionT {
	return VersionT{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
	}
}

func (v VersionT) String() string {
	return fmt.Sprintf("Version: %s, Commit: %s, BuildDate: %s", v.Version, v.Commit, v.BuildDate)
}

func (v VersionT) UserAgent() string {
	return fmt.Sprintf("escape-cli/%s", v.Version)
}
