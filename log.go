package main

import (
	"fmt"
	"io"
	"os"
)

// build
var logWriter = io.Writer(os.Stderr)

// common
func logPrint(s string) {
	fmt.Fprint(logWriter, s)
}

func logPrintf(f string, v ...any) {
	fmt.Fprintf(logWriter, f, v...)
}

// Info
func logInfo(s string) {
	logPrint("I: " + s)
}

func logInfoln(s string) {
	logInfo(s + "\n")
}

func logInfof(f string, v ...any) {
	s := fmt.Sprintf(f, v...)
	logInfo(s)
}

// Error
func logError(s string) {
	logPrint("E: " + s)
}

func logErrorln(s string) {
	logError(s + "\n")
}

func logErrorf(f string, v ...any) {
	s := fmt.Sprintf(f, v...)
	logError(s)
}

// Warning
func logWarning(s string) {
	logPrint("W: " + s)
}

func logWarningln(s string) {
	logWarning(s + "\n")
}

func logWarningf(f string, v ...any) {
	s := fmt.Sprintf(f, v...)
	logError(s)
}
