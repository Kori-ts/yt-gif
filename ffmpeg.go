package main

import (
	"fmt"
	"strconv"
	"strings"
)

func gifFilter(opts Options) string {
	var filters []string
	if opts.FPS != "source" {
		filters = append(filters, "fps="+opts.FPS)
	}
	if opts.Width > 0 || opts.Height > 0 {
		filters = append(filters, scaleFilter(opts))
	}

	prefix := ""
	if len(filters) > 0 {
		prefix = strings.Join(filters, ",") + ","
	}
	return prefix + "split[s0][s1];[s0]palettegen=stats_mode=diff[p];[s1][p]paletteuse=dither=sierra2_4a"
}

func scaleFilter(opts Options) string {
	width := "-1"
	height := "-1"
	if opts.Width > 0 {
		width = strconv.Itoa(opts.Width)
	}
	if opts.Height > 0 {
		height = strconv.Itoa(opts.Height)
	}
	return fmt.Sprintf("scale=%s:%s:flags=lanczos", width, height)
}
