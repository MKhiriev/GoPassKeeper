package models

type AppBuildInfo struct {
	buildVersion string
	buildDate    string
	buildCommit  string
}

func NewAppBuildInfo(buildVersion, buildDate, buildCommit string) AppBuildInfo {
	return AppBuildInfo{
		buildVersion: buildVersion,
		buildDate:    buildDate,
		buildCommit:  buildCommit,
	}
}

func (a AppBuildInfo) BuildVersion() string {
	return a.buildVersion
}

func (a AppBuildInfo) BuildDate() string {
	return a.buildDate
}

func (a AppBuildInfo) BuildCommit() string {
	return a.buildCommit
}
