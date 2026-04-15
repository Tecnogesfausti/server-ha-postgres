package version

import "strconv"

const notSet string = "1.29.3-tecnoges-incoming.3"

// This information will be collected when build, by `-ldflags "-X main.appVersion=0.1"`.
//
//nolint:gochecknoglobals // build-time constant
var (
	AppVersion = notSet
	AppRelease = "20260415"
)

func AppReleaseID() int {
	id, _ := strconv.Atoi(AppRelease)

	return id
}
