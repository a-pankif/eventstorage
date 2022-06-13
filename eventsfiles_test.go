package eventstorage

import (
	"bytes"
	"strings"
	"testing"
)

func Test_eventStorage_initRegistryFile(t *testing.T) {
	b := &eventStorage{
		basePath: t.TempDir(),
		write:    &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:     &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}

	t.Cleanup(b.Shutdown)

	if err := b.initRegistryFile(); err != nil {
		t.Errorf("filesRegistry not inited")
		return
	}
}

func Test_eventStorage_initRegistryFileFailed(t *testing.T) {
	b := &eventStorage{
		basePath: string([]byte{0}),
		write:    &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:     &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}
	t.Cleanup(b.Shutdown)

	if err := b.initRegistryFile(); err == nil {
		t.Errorf("filesRegistry inited, but expected fail")
	}
}

func Test_eventStorage_appendInRegistryFile(t *testing.T) {
	b := &eventStorage{
		basePath: t.TempDir(),
		write:    &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:     &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}

	_ = b.initRegistryFile()

	file1, _ := b.openEventsFile(1, true)
	file2, _ := b.openEventsFile(2, true)

	_ = file1.Close()
	_ = file2.Close()
	_ = b.filesRegistry.Close()

	_ = b.initRegistryFile()

	// expectedMap := logFilesMap{1: "events.1", 2: "events.2"}
	// isEqual := reflect.DeepEqual(b.eventsFilesMap, expectedMap)
	//
	// if !isEqual {
	// 	t.Errorf("eventsFilesMap not equal expected")
	// }
	// todo - repair

	t.Cleanup(b.Shutdown)
}

func Test_eventStorage_initLogFileWithoutRegistry(t *testing.T) {
	b := &eventStorage{
		write: &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:  &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}

	if err := b.initEventsFile(); err == nil {
		t.Errorf("initEventsFile expected failed without registry")
	}

	t.Cleanup(b.Shutdown)
}

func Test_eventStorage_initLogFile(t *testing.T) {
	b := &eventStorage{
		basePath: t.TempDir(),
		write:    &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:     &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}
	_ = b.initRegistryFile()

	if err := b.initEventsFile(); err != nil {
		t.Errorf("initEventsFile failed")
		return
	}

	t.Cleanup(b.Shutdown)
}

func Test_eventStorage_initLogFileFailed(t *testing.T) {
	b := &eventStorage{
		basePath: t.TempDir(),
		write:    &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:     &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}

	_ = b.initRegistryFile()

	t.Cleanup(b.Shutdown)

	b.basePath = string([]byte{0})

	if err := b.initEventsFile(); err == nil {
		t.Errorf("initEventsFile expect failed")
	}
}

func Test_eventStorage_rotateLogFileFailedCloseOld(t *testing.T) {
	b := &eventStorage{
		write: &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:  &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}

	if err := b.rotateEventsFile(); err == nil {
		t.Errorf("rotateLogFileFailedCloseOld expect failed")
	}

	t.Cleanup(b.Shutdown)
}

func Test_eventStorage_rotateLogFile(t *testing.T) {
	b := &eventStorage{
		basePath: t.TempDir(),
		write:    &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:     &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}

	_ = b.initRegistryFile()
	_ = b.initEventsFile()

	t.Cleanup(b.Shutdown)

	if err := b.rotateEventsFile(); err != nil {
		t.Errorf("rotateEventsFile failed")
		return
	}
	// todo - repair
	// expectedMap := logFilesMap{1: "events.1", 2: "events.2"}
	// isEqual := reflect.DeepEqual(b.eventsFilesMap, expectedMap)
	//
	// if !isEqual {
	// 	t.Errorf("rotateEventsFile eventsFilesMap not equal expected")
	// }
}

func Test_eventStorage_openLogFileFailedAppend(t *testing.T) {
	b := &eventStorage{
		write: &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:  &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}
	t.Cleanup(b.Shutdown)

	if _, err := b.openEventsFile(1, true); err == nil {
		t.Errorf("openLogFileFailedAppend expect failed")
	}
}

func Test_eventStorage_SetLogFileSize(t *testing.T) {
	b := &eventStorage{
		write: &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:  &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}
	b.SetLogFileMaxSize(100)
	t.Cleanup(b.Shutdown)

	if b.write.fileMaxSize != 100 {
		t.Errorf("SetLogFileMaxSize failed")
	}
}

func Test_eventStorage_calculateLogFileSize(t *testing.T) {
	b := &eventStorage{
		basePath: t.TempDir(),
		write:    &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:     &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}

	_ = b.initRegistryFile()
	_ = b.initEventsFile()

	t.Cleanup(b.Shutdown)

	_, _ = b.write.file.Write([]byte{1, 2, 3})

	if b.calculateLogFileSize() != 3 {
		t.Errorf("calculateLogFileSize failed")
	}
}
