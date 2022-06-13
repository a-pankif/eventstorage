package eventstorage

import (
	"bytes"
	"errors"
	"os"
	"strings"
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
	readEventsOpLimit       = 100
)

var (
	ErrAutoFlushTimeAlreadySet = errors.New("autoFlushTime already set")
	ErrAutoFlushTimeTooLow     = errors.New("autoFlushTime too low value")
	eventsFileNameTemplate     = "events.%d"
)

type eventStorage struct {
	basePath      string   // Root path of events storage.
	filesRegistry *os.File // File with list of exists events files.
	write         *write   // Variables for write events.
	read          *read    // Variables for read events.
	turnedOff     chan bool
}

type write struct {
	file           *os.File      // Current file to write events
	fileSize       int64         // Size of current events file
	fileMaxSize    int64         // Size of events file for create a new file
	locker         sync.Mutex    // Write common variables lock to avoid race condition.
	buf            *bytes.Buffer // For collect data before flush it to file.
	insertsCount   int           // Count of written events, from last data flush
	autoFlushCount int           // Auto flush after N count of events insert, 0 - disable.
	autoFlushTime  time.Duration // Auto flush every N seconds, 0 - disable.

}

type read struct {
	locker        sync.Mutex       // Read common variables lock to avoid race condition.
	buf           *strings.Builder // For collect event data before append to events slice.
	seekOffset    int64            // Current file read offset.
	eventsSaved   int              // Count of events to return.
	eventsCount   int              // Count of read events, for check offset.
	readableFiles readableFiles    // Map of events files opened for read.
}

type readableFiles map[int]*os.File
