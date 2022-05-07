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

type blogger struct {
	logFile        *os.File
	errWriter      io.Writer
	logWriter      io.Writer
	buf            *bytes.Buffer
	bufLock        sync.Mutex
	insertsCount   int           // Count of logged events
	autoFlushCount int           // Auto flush after N count of log insert, 0 - disable.
	autoFlushTime  time.Duration // Auto flush every N seconds, 0 - disable.
	lastLineUsed   int           // Number of bytes in the current line.
}
