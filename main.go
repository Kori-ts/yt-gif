package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const usageText = `yt-gif is a simple CLI tool for creating GIFs from YouTube videos.

Usage:
  yt-gif check
  yt-gif <url> <start-end>... [options]

Options:
  --out <dir>          output directory. Defaults to the current directory.
  --force              overwrite existing clip-{n}.gif files.
  --jobs <n>           parallel jobs. Defaults to 2.
  --fps <n|source>     GIF framerate. Must be a numeric value in the range 1-60. Defaults to 15.
  --width <px|source>  output width. Use source to leave width unconstrained. Defaults to 640.
  --height <px|source> output height. Use source to leave height unconstrained.
  --timing             show download, conversion, and total elapsed times.
  --accurate-cut       force exact section cuts in yt-dlp. Slower.
  --verbose            show yt-dlp and ffmpeg output with job prefixes.
  --help               show this help.
`

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

func run(args []string, log io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(log, usageText)
		return 1
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprint(log, usageText)
		return 0
	}
	if args[0] == "check" {
		return runCheck(log)
	}

	cfg, err := parseCommand(args)
	if err != nil {
		return fail(log, "%v", err)
	}

	missing := missingDependencies()
	if len(missing) > 0 {
		for _, name := range missing {
			errorf(log, "missing dependency in PATH: %s", name)
		}
		return 1
	}

	ranges := parseValidRanges(cfg.RangeArgs, log)
	if len(ranges) == 0 {
		return fail(log, "no valid timestamp ranges")
	}

	outDir, err := createOutputDir(cfg.Options.OutDir)
	if err != nil {
		return fail(log, "%v", err)
	}

	jobs, err := planJobs(cfg.URL, ranges, outDir, cfg.Options.Force)
	if err != nil {
		return fail(log, "%v", err)
	}

	runTmp, err := os.MkdirTemp("", "yt-gif-*")
	if err != nil {
		return fail(log, "could not create temp directory: %v", err)
	}
	defer os.RemoveAll(runTmp)
	for i := range jobs {
		jobs[i].TempDir = filepath.Join(runTmp, fmt.Sprintf("job-%d", jobs[i].Index))
	}

	results := runJobs(jobs, cfg.Options, log)
	ok := 0
	for _, res := range results {
		if res.Err != nil {
			warnf(log, "clip-%d.gif  %s failed: %v", res.Index, jobs[res.Index-1].Range.Raw, res.Err)
			continue
		}
		ok++
	}
	if ok == 0 {
		return fail(log, "no GIFs were created")
	}
	return 0
}

func fail(log io.Writer, format string, args ...any) int {
	errorf(log, format, args...)
	return 1
}

func parseValidRanges(rawRanges []string, log io.Writer) []ClipRange {
	ranges := make([]ClipRange, 0, len(rawRanges))
	for _, raw := range rawRanges {
		rng, err := parseRange(raw)
		if err != nil {
			warnf(log, "skipping invalid range %q: %v", raw, err)
			continue
		}
		ranges = append(ranges, rng)
	}
	return ranges
}

func createOutputDir(raw string) (string, error) {
	outDir, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("invalid output path: %w", err)
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", fmt.Errorf("could not create output directory: %w", err)
	}
	return outDir, nil
}
