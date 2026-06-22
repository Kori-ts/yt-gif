package main

import (
	"io"
	"os/exec"
)

var dependencyNames = []string{"yt-dlp", "ffmpeg"}

func runCheck(log io.Writer) int {
	missing := false
	for _, name := range dependencyNames {
		path, err := exec.LookPath(name)
		if err != nil {
			errorf(log, "missing dependency in PATH: %s", name)
			missing = true
			continue
		}
		okf(log, "%s found: %s", name, path)
	}
	if missing {
		return 1
	}
	return 0
}

func missingDependencies() []string {
	var missing []string
	for _, name := range dependencyNames {
		if _, err := exec.LookPath(name); err != nil {
			missing = append(missing, name)
		}
	}
	return missing
}
