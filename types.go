package binarylog

import (
	"bytes"
	"errors"
	"io"
	"os"
	"sync"
	"time"
)

const (
	lineLength = 40
)

var (
	ErrAutoFlushTimeAlreadySet = errors.New("autoFlushTime already set")
	ErrAutoFlushTimeTooLow     = errors.New("autoFlushTime too low value")
)

type binaryLogger struct {
	logFile            *os.File
	errWriter          io.Writer
	logWriter          io.Writer
	buf                *bytes.Buffer // For collect encoded data before flush it to file.
	encodeBuf          []byte        // 2 bytes slice for HEX encoding, 1 byte for space or line break.
	bufLock            sync.Mutex    // Lock buf and encodeBuf vars.
	insertsCount       int           // Count of logged events.
	autoFlushCount     int           // Auto flush after N count of log insert, 0 - disable.
	autoFlushTime      time.Duration // Auto flush every N seconds, 0 - disable.
	lastLineBytesCount int           // Number of bytes in the last line: only pure bytes (not hex encoded), without spaces and line breaks.
}
