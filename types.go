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
	ErrLogNotInited            = errors.New("log init failed")
	logFileTemplate            = "binlog.%d"
	RowDelimiter               = []byte{EmptyByte, EmptyByte, EmptyByte, EmptyByte, EmptyByte, EmptyByte, EmptyByte, EmptyByte}
)

type binaryLogger struct {
	basePath           string
	currentLogFile     *os.File
	logFilesRegistry   *os.File
	logFilesCount      int   // Count of created log files
	logFileMaxSize     int64 // Size of log file for create a new file
	currenLogFileSize  int64
	errWriter          io.Writer
	logWriter          io.Writer
	buf                *bytes.Buffer // For collect encoded data before flush it to file.
	encodeBuf          []byte        // 2 bytes slice for HEX encoding, 1 byte for Space or line break.
	bufLock            sync.Mutex    // Lock buf and encodeBuf vars.
	insertsCount       int           // Count of logged events.
	autoFlushCount     int           // Auto flush after N count of log insert, 0 - disable.
	autoFlushTime      time.Duration // Auto flush every N seconds, 0 - disable.
	lastLineBytesCount int           // Number of bytes in the last line: only pure bytes (not hex encoded), without spaces and line breaks.
}
