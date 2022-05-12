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
	lineLength            = 40
	Space            byte = ' '
	LineBreak        byte = '\n'
	EmptyByte        byte = 0
	registryFileName      = "binlog.registry"
)

const (
	KB int64 = 1 << 10
	MB int64 = 1 << 20
)

var (
	ErrAutoFlushTimeAlreadySet = errors.New("autoFlushTime already set")
	ErrAutoFlushTimeTooLow     = errors.New("autoFlushTime too low value")
	logFileTemplate            = "binlog.%d"
	RowDelimiter               = []byte{EmptyByte, EmptyByte, EmptyByte, EmptyByte, EmptyByte, EmptyByte, EmptyByte, EmptyByte}
)

type binaryLogger struct {
	basePath           string
	logFile            *os.File      // Current log file to write logs
	logFilesRegistry   *os.File      // File with list of exists log files
	logFilesMap        logFilesMap   // Map presentation of log files registry
	logFilesCount      int           // Count of created log files
	logFileMaxSize     int64         // Size of log file for create a new file
	logFileSize        int64         // Size of current log file
	buf                *bytes.Buffer // For collect encoded data before flush it to file.
	encodeBuf          []byte        // 2 bytes slice for HEX encoding, 1 byte for Space or LineBreak.
	bufLock            sync.Mutex    // Lock buf and encodeBuf vars.
	insertsCount       int           // Count of logged events, from last data flush
	autoFlushCount     int           // Auto flush after N count of log insert, 0 - disable.
	autoFlushTime      time.Duration // Auto flush every N seconds, 0 - disable.
	lastLineBytesCount int           // Number of bytes in the last line: only pure bytes (not hex encoded), without spaces and line breaks.

	errWriter io.Writer
	logWriter io.Writer
}

type logFilesMap map[int]string
