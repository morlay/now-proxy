package version

const (
	DevelopmentVersion = "devel"
)

var (
	Version = DevelopmentVersion
)

func FullVersion() string {
	return Version
}
