// Package buildinfo reports release and module build metadata.
package buildinfo

import (
	"fmt"
	"runtime/debug"
	"strings"
)

var (
	version   = "devel"
	commit    string
	buildDate string
)

// Info contains user-visible build metadata.
type Info struct {
	Version   string
	Commit    string
	BuildDate string
}

// Current returns linker-injected metadata with Go module/VCS fallback.
func Current() Info {
	info := Info{Version: version, Commit: commit, BuildDate: buildDate}
	build, ok := debug.ReadBuildInfo()
	if !ok {
		return info
	}
	if (info.Version == "" || info.Version == "devel") && build.Main.Version != "" && build.Main.Version != "(devel)" {
		info.Version = build.Main.Version
	}
	for _, setting := range build.Settings {
		switch setting.Key {
		case "vcs.revision":
			if info.Commit == "" {
				info.Commit = setting.Value
			}
		case "vcs.time":
			if info.BuildDate == "" {
				info.BuildDate = setting.Value
			}
		}
	}
	return info
}

// String returns stable single-line version output.
func (i Info) String() string {
	versionValue := i.Version
	if versionValue == "" {
		versionValue = "devel"
	}
	if versionValue != "devel" && !strings.HasPrefix(versionValue, "v") {
		versionValue = "v" + versionValue
	}
	var metadata []string
	if i.Commit != "" {
		commitValue := i.Commit
		if len(commitValue) > 7 {
			commitValue = commitValue[:7]
		}
		metadata = append(metadata, "commit "+commitValue)
	}
	if i.BuildDate != "" {
		metadata = append(metadata, "built "+i.BuildDate)
	}
	if len(metadata) == 0 {
		return "papercuts " + versionValue
	}
	return fmt.Sprintf("papercuts %s (%s)", versionValue, strings.Join(metadata, ", "))
}
