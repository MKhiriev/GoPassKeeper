// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package models

// AppBuildInfo carries immutable build-time metadata embedded into binaries.
//
// Values are typically injected by linker flags during CI/CD and shown in
// CLI/TUI version output for diagnostics and release traceability.
type AppBuildInfo struct {
	buildVersion string
	buildDate    string
	buildCommit  string
}

// NewAppBuildInfo constructs [AppBuildInfo] from the provided build metadata.
func NewAppBuildInfo(buildVersion, buildDate, buildCommit string) AppBuildInfo {
	return AppBuildInfo{
		buildVersion: buildVersion,
		buildDate:    buildDate,
		buildCommit:  buildCommit,
	}
}

// BuildVersion returns the semantic version string of the build.
func (a AppBuildInfo) BuildVersion() string {
	return a.buildVersion
}

// BuildDate returns the build timestamp string.
func (a AppBuildInfo) BuildDate() string {
	return a.buildDate
}

// BuildCommit returns the source-control commit hash used for the build.
func (a AppBuildInfo) BuildCommit() string {
	return a.buildCommit
}
