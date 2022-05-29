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
	LineBreak        byte = '\n'
	registryFileName      = "events_files.registry"
)

const (
	KB                int64 = 1 << 10
	MB                int64 = 1 << 20
	readEventsOpLimit       = 1
)

var (
	ErrAutoFlushTimeAlreadySet = errors.New("autoFlushTime already set")
	ErrAutoFlushTimeTooLow     = errors.New("autoFlushTime too low value")
	ErrEventsFileNotExists     = errors.New("cant find events file")
	eventsFileNameTemplate     = "events.%d"
)

type eventStorage struct {
	basePath            string
	eventsFile          *os.File    // Current file to write events
	eventsFileSize      int64       // Size of current events file
	eventsFilesRegistry *os.File    // File with list of exists events files
	eventsFilesMap      logFilesMap // Map representation of events files registry
	eventsFilesReadMap  eventsFileRead
	filesCount          int           // Count of created events files
	fileMaxSize         int64         // Size of events file for create a new file
	buf                 *bytes.Buffer // For collect data before flush it to file.
	writeLocker         sync.Mutex    // Write common variables lock to avoid race condition.
	readLocker          sync.Mutex    // Read common variables lock to avoid race condition.
	insertsCount        int           // Count of written events, from last data flush
	autoFlushCount      int           // Auto flush after N count of events insert, 0 - disable.
	autoFlushTime       time.Duration // Auto flush every N seconds, 0 - disable.

	errWriter io.Writer
}

type logFilesMap map[int]string
type eventsFileRead map[int]*os.File
