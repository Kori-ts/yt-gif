package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Job struct {
	Index   int
	URL     string
	Range   ClipRange
	Output  string
	TempDir string
}

type JobResult struct {
	Index  int
	Output string
	Err    error
}

func planJobs(url string, ranges []ClipRange, outDir string, force bool) ([]Job, error) {
	jobs := make([]Job, 0, len(ranges))
	for i, rng := range ranges {
		out := filepath.Join(outDir, fmt.Sprintf("clip-%d.gif", i+1))
		if !force {
			if _, err := os.Stat(out); err == nil {
				return nil, fmt.Errorf("output already exists: %s", out)
			} else if !errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("could not check output path %s: %w", out, err)
			}
		}
		jobs = append(jobs, Job{
			Index:  i + 1,
			URL:    url,
			Range:  rng,
			Output: out,
		})
	}
	return jobs, nil
}

func runJobs(jobs []Job, opts Options, log io.Writer) []JobResult {
	if len(jobs) == 0 { return nil }

	workers := opts.Jobs
	if workers > len(jobs) { workers = len(jobs) }

	ctx := context.Background()
	in := make(chan Job)
	out := make(chan JobResult)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range in {
				out <- processJob(ctx, j, opts, log)
			}
		}()
	}

	go func() {
		for _, j := range jobs {
			in <- j
		}
		close(in)
		wg.Wait()
		close(out)
	}()

	results := make([]JobResult, len(jobs))
	for res := range out {
		results[res.Index-1] = res
	}
	return results
}

func processJob(ctx context.Context, j Job, opts Options, log io.Writer) JobResult {
	name := filepath.Base(j.Output)
	start := time.Now()

	infof(log, "%s  %s  downloading section", name, j.Range.Raw)
	if err := os.MkdirAll(j.TempDir, 0755); err != nil {
		return JobResult{Index: j.Index, Output: j.Output, Err: err}
	}

	if err := downloadSection(ctx, j, opts, log); err != nil {
		return JobResult{Index: j.Index, Output: j.Output, Err: fmt.Errorf("yt-dlp failed: %w", err)}
	}
	if opts.Timing { infof(log, "%s  %s  downloaded section in %s", name, j.Range.Raw, elapsed(start)) }

	sectionFile, err := findSectionFile(j.TempDir)
	if err != nil { return JobResult{Index: j.Index, Output: j.Output, Err: err} }

	convertStart := time.Now()
	infof(log, "%s  %s  converting gif", name, j.Range.Raw)
	if err := convertSection(ctx, sectionFile, j, opts, log); err != nil {
		return JobResult{Index: j.Index, Output: j.Output, Err: fmt.Errorf("ffmpeg failed: %w", err)}
	}
	if opts.Timing {
		infof(log, "%s  %s  converted gif in %s", name, j.Range.Raw, elapsed(convertStart))
		infof(log, "%s  %s  done in %s", name, j.Range.Raw, elapsed(start))
	} else {
		infof(log, "%s  %s  done", name, j.Range.Raw)
	}
	return JobResult{Index: j.Index, Output: j.Output}
}

func downloadSection(ctx context.Context, j Job, opts Options, log io.Writer) error {
	section := "*" + j.Range.Start + "-" + j.Range.End
	args := []string{
		j.URL,
		"-f", "bestvideo/best",
	}
	if sort := formatSort(opts); sort != "" { args = append(args, "-S", sort) }
	if opts.AccurateCut { args = append(args, "--force-keyframes-at-cuts") }
	args = append(args,
		"--no-playlist",
		"--download-sections", section,
		"--paths", "home:" + j.TempDir,
		"--paths", "temp:" + j.TempDir,
		"-o", "section.%(ext)s",
	)
	return runCmd(ctx, opts.Verbose, fmt.Sprintf("clip-%d yt-dlp", j.Index), "yt-dlp", args, log)
}

func formatSort(opts Options) string {
	var fields []string
	if opts.Width > 0 { fields = append(fields, fmt.Sprintf("width:%d", opts.Width)) }
	if opts.Height > 0 { fields = append(fields, fmt.Sprintf("height:%d", opts.Height)) }
	return strings.Join(fields, ",")
}

func elapsed(start time.Time) time.Duration {
	return time.Since(start).Round(time.Millisecond)
}

func convertSection(ctx context.Context, sectionFile string, j Job, opts Options, log io.Writer) error {
	args := []string{
		"-y",
	}
	if !opts.Verbose { args = append(args, "-hide_banner", "-loglevel", "error", "-nostats") }
	args = append(args,
		"-i", sectionFile,
		"-filter_complex", gifFilter(opts),
		"-gifflags", "-offsetting",
		j.Output,
	)
	return runCmd(ctx, opts.Verbose, fmt.Sprintf("clip-%d ffmpeg", j.Index), "ffmpeg", args, log)
}

func findSectionFile(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil { return "", err }
	var files []string
	for _, entry := range entries {
		if entry.IsDir() { continue }
		name := entry.Name()
		if strings.HasPrefix(name, "section.") && !strings.HasSuffix(name, ".part") && !strings.HasSuffix(name, ".ytdl") {
			files = append(files, filepath.Join(dir, name))
		}
	}
	if len(files) == 0 { return "", errors.New("yt-dlp did not produce a section file") }
	if len(files) > 1 { return "", fmt.Errorf("yt-dlp produced multiple section files in %s", dir) }
	return files[0], nil
}

func runCmd(ctx context.Context, verbose bool, prefix, name string, args []string, log io.Writer) error {
	cmd := exec.CommandContext(ctx, name, args...)
	var buf bytes.Buffer
	if verbose {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}
		if err := cmd.Start(); err != nil {
			return err
		}
		var wg sync.WaitGroup
		wg.Add(2)
		go prefixLines(&wg, log, prefix, stdout)
		go prefixLines(&wg, log, prefix, stderr)
		wg.Wait()
		return cmd.Wait()
	}

	cmd.Stdout = io.Discard
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		msg := lastUsefulLine(buf.String())
		if msg == "" { return err }
		return fmt.Errorf("%v: %s", err, msg)
	}
	return nil
}

func prefixLines(wg *sync.WaitGroup, log io.Writer, prefix string, r io.Reader) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		fmt.Fprintf(log, "[%s] %s\n", prefix, scanner.Text())
	}
}

func lastUsefulLine(s string) string {
	lines := strings.Split(strings.ReplaceAll(s, "\r", "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" { return line }
	}
	return ""
}
