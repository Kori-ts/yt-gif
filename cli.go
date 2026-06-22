package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Options struct {
	OutDir      string
	Force       bool
	Jobs        int
	FPS         string
	Width       int
	Height      int
	Timing      bool
	Verbose     bool
	AccurateCut bool
}

type CommandConfig struct {
	URL       string
	RangeArgs []string
	Options   Options
}

func parseCommand(args []string) (CommandConfig, error) {
	if len(args) == 0 {
		return CommandConfig{}, errors.New("missing URL")
	}

	opts, ranges, err := parseOptions(args[1:])
	if err != nil {
		return CommandConfig{}, err
	}
	if len(ranges) == 0 {
		return CommandConfig{}, errors.New("at least one timestamp range is required")
	}

	return CommandConfig{
		URL:       args[0],
		RangeArgs: ranges,
		Options:   opts,
	}, nil
}

func parseOptions(args []string) (Options, []string, error) {
	opts := Options{OutDir: ".", Jobs: 2, FPS: "15", Width: 640}
	var ranges []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--out":
			value, next, err := optionValue(args, i, "--out")
			if err != nil {
				return opts, nil, err
			}
			opts.OutDir = value
			i = next
		case "--force":
			opts.Force = true
		case "--jobs":
			value, next, err := optionValue(args, i, "--jobs")
			if err != nil {
				return opts, nil, err
			}
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 {
				return opts, nil, errors.New("--jobs must be a positive integer")
			}
			opts.Jobs = n
			i = next
		case "--fps":
			value, next, err := optionValue(args, i, "--fps")
			if err != nil {
				return opts, nil, err
			}
			fps, err := parseFPS(value)
			if err != nil {
				return opts, nil, err
			}
			opts.FPS = fps
			i = next
		case "--width":
			value, next, err := optionValue(args, i, "--width")
			if err != nil {
				return opts, nil, err
			}
			width, err := parseDimension(value, "--width")
			if err != nil {
				return opts, nil, err
			}
			opts.Width = width
			i = next
		case "--height":
			value, next, err := optionValue(args, i, "--height")
			if err != nil {
				return opts, nil, err
			}
			height, err := parseDimension(value, "--height")
			if err != nil {
				return opts, nil, err
			}
			opts.Height = height
			i = next
		case "--verbose":
			opts.Verbose = true
		case "--timing":
			opts.Timing = true
		case "--accurate-cut":
			opts.AccurateCut = true
		case "--help", "-h":
			return opts, nil, errors.New(usageText)
		default:
			if strings.HasPrefix(arg, "--") {
				return opts, nil, fmt.Errorf("unknown option: %s", arg)
			}
			ranges = append(ranges, arg)
		}
	}
	return opts, ranges, nil
}

func optionValue(args []string, index int, name string) (string, int, error) {
	next := index + 1
	if next >= len(args) {
		return "", index, fmt.Errorf("%s requires a value", name)
	}
	return args[next], next, nil
}

func parseFPS(value string) (string, error) {
	if value == "source" {
		return value, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 || n > 60 {
		return "", errors.New("--fps must be source or an integer from 1 to 60")
	}
	return strconv.Itoa(n), nil
}

func parseDimension(value, name string) (int, error) {
	if value == "source" {
		return 0, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("%s must be source or a positive integer", name)
	}
	return n, nil
}
