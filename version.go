package main

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	// Major version
	Major = 0

	// Minor version
	Minor = 1

	// PatchSet version
	PatchSet = 0
)

var (
	// ErrBuildShaNotSet is returned if the build sha was not injected.
	ErrBuildShaNotSet = errors.New("build sha not set")

	// ErrBuildTimeNotSet is returned if the build time was not injected.
	ErrBuildTimeNotSet = errors.New("build time not set")
)

var (
	// BuildSha is a git commit sha injected in build time.
	BuildSha string

	// BuildTime contains a time the binary was built. Injected in build time.
	BuildTime string
)

// NewVersion returns a new instance of version object.
func NewVersion() (*Version, error) {
	if BuildSha == "" {
		return nil, ErrBuildShaNotSet
	}

	if BuildTime == "" {
		return nil, ErrBuildTimeNotSet
	}

	return &Version{
		Major:     Major,
		Minor:     Minor,
		PatchSet:  PatchSet,
		BuildSha:  BuildSha,
		BuildTime: BuildTime,
	}, nil
}

// Version represents an application version, which includes major, minor and patchset versions
// also build sha and build time.
type Version struct {
	Major     int    `json:"major"`
	Minor     int    `json:"minor"`
	PatchSet  int    `json:"patch_set"`
	BuildSha  string `json:"build_sha"`
	BuildTime string `json:"build_time"`
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d - %s ; built on %s", Major, Minor, PatchSet, BuildSha, BuildTime)
}
