package main

import (
	"fmt"
	"io"
)

func okf(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "[OK] "+format+"\n", args...)
}

func infof(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "[INFO] "+format+"\n", args...)
}

func warnf(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "[WARNING] "+format+"\n", args...)
}

func errorf(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "[ERROR] "+format+"\n", args...)
}
