package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type ClipRange struct {
	Raw   string
	Start string
	End   string
}

func parseRange(raw string) (ClipRange, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ClipRange{}, errors.New("empty range")
	}
	if strings.Count(s, "-") != 1 {
		return ClipRange{}, errors.New("range must have exactly one '-' separator")
	}

	parts := strings.SplitN(s, "-", 2)
	start := parts[0]
	end := parts[1]
	if start == "" || end == "" {
		return ClipRange{}, errors.New("start and end must be non-empty")
	}
	if hasControlOrSpace(start) || hasControlOrSpace(end) {
		return ClipRange{}, errors.New("timestamps cannot contain whitespace or control characters")
	}

	startDur, err := parseTimestamp(start)
	if err != nil {
		return ClipRange{}, fmt.Errorf("invalid start timestamp: %v", err)
	}
	endDur, err := parseTimestamp(end)
	if err != nil {
		return ClipRange{}, fmt.Errorf("invalid end timestamp: %v", err)
	}
	if endDur <= startDur {
		return ClipRange{}, errors.New("end must be greater than start")
	}

	return ClipRange{Raw: s, Start: start, End: end}, nil
}

func hasControlOrSpace(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

func parseTimestamp(s string) (time.Duration, error) {
	parts := strings.Split(s, ":")
	if len(parts) > 3 {
		return 0, errors.New("too many ':' separators")
	}

	switch len(parts) {
	case 1:
		sec, nanos, err := parseSeconds(parts[0])
		if err != nil {
			return 0, err
		}
		return time.Duration(sec)*time.Second + time.Duration(nanos), nil
	case 2:
		min, err := parseWholeNumber(parts[0], "minutes")
		if err != nil {
			return 0, err
		}
		sec, nanos, err := parseSeconds(parts[1])
		if err != nil {
			return 0, err
		}
		if sec > 59 {
			return 0, errors.New("seconds must be 0-59")
		}
		return time.Duration(min*60+sec)*time.Second + time.Duration(nanos), nil
	case 3:
		hour, err := parseWholeNumber(parts[0], "hours")
		if err != nil {
			return 0, err
		}
		min, err := parseWholeNumber(parts[1], "minutes")
		if err != nil {
			return 0, err
		}
		if min > 59 {
			return 0, errors.New("minutes must be 0-59")
		}
		sec, nanos, err := parseSeconds(parts[2])
		if err != nil {
			return 0, err
		}
		if sec > 59 {
			return 0, errors.New("seconds must be 0-59")
		}
		return time.Duration(hour*3600+min*60+sec)*time.Second + time.Duration(nanos), nil
	default:
		return 0, errors.New("invalid timestamp")
	}
}

func parseWholeNumber(s, label string) (int64, error) {
	if s == "" {
		return 0, fmt.Errorf("%s cannot be empty", label)
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("%s must be numeric", label)
		}
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s is too large", label)
	}
	return n, nil
}

func parseSeconds(s string) (int64, int64, error) {
	if s == "" {
		return 0, 0, errors.New("seconds cannot be empty")
	}
	if strings.Count(s, ".") > 1 {
		return 0, 0, errors.New("seconds can contain at most one decimal point")
	}

	whole := s
	frac := ""
	if strings.Contains(s, ".") {
		parts := strings.SplitN(s, ".", 2)
		whole = parts[0]
		frac = parts[1]
		if frac == "" {
			return 0, 0, errors.New("fraction cannot be empty")
		}
	}

	sec, err := parseWholeNumber(whole, "seconds")
	if err != nil {
		return 0, 0, err
	}
	nanos, err := parseFraction(frac)
	if err != nil {
		return 0, 0, err
	}
	return sec, nanos, nil
}

func parseFraction(frac string) (int64, error) {
	if frac == "" {
		return 0, nil
	}
	if len(frac) > 9 {
		frac = frac[:9]
	}
	for _, r := range frac {
		if r < '0' || r > '9' {
			return 0, errors.New("fraction must be numeric")
		}
	}
	for len(frac) < 9 {
		frac += "0"
	}
	nanos, err := strconv.ParseInt(frac, 10, 64)
	if err != nil {
		return 0, errors.New("fraction is too large")
	}
	return nanos, nil
}
