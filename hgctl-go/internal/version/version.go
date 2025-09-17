package version

var (
	Version string
	Commit  string
)

func GetVersion() string {
	return Version
}

func GetCommit() string {
	return Commit
}

func GetFullVersion() string {
	return Version + " (" + Commit + ")"
}
