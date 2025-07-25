package version

var (
    Version = "dev"
    Commit  = "unknown"
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
