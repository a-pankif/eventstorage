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

	if err := b.initFilesRegistry(); err != nil {
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

	if err := b.initFilesRegistry(); err == nil {
		t.Errorf("filesRegistry inited, but expected fail")
	}
}

func Test_eventStorage_appendInRegistryFile(t *testing.T) {
	b := &eventStorage{
		basePath: t.TempDir(),
		write:    &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:     &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}
	t.Cleanup(b.Shutdown)
	_ = b.initFilesRegistry()

	file1, _ := b.openEventsFile(1, true)
	file2, _ := b.openEventsFile(2, true)
	file3, _ := b.openEventsFile(3, true)

	_ = file1.Close()
	_ = file2.Close()
	_ = file3.Close()
	_ = b.filesRegistry.Close()

	if len(b.read.readableFiles) != 3 {
		t.Errorf("appendInRegistryFile has wrong count")
		return
	}

	for i, file := range b.read.readableFiles {
		info, _ := file.Stat()

		if b.getFileName(i) != info.Name() {
			t.Errorf("Wrong file name, expected %v, got %v", b.getFileName(i), info.Name())
			return
		}
	}
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
	_ = b.initFilesRegistry()

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

	_ = b.initFilesRegistry()

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

	_ = b.initFilesRegistry()
	_ = b.initEventsFile()

	t.Cleanup(b.Shutdown)

	if err := b.rotateEventsFile(); err != nil {
		t.Errorf("rotateEventsFile failed")
		return
	}

	if len(b.read.readableFiles) != 2 {
		t.Errorf("rotateLogFile has wrong count")
		return
	}

	for i, file := range b.read.readableFiles {
		info, _ := file.Stat()

		if b.getFileName(i) != info.Name() {
			t.Errorf("Wrong file name, expected %v, got %v", b.getFileName(i), info.Name())
			return
		}
	}
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
	b.SetWriteFileMaxSize(100)
	t.Cleanup(b.Shutdown)

	if b.write.fileMaxSize != 100 {
		t.Errorf("SetWriteFileMaxSize failed")
	}
}

func Test_eventStorage_calculateLogFileSize(t *testing.T) {
	b := &eventStorage{
		basePath: t.TempDir(),
		write:    &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:     &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}

	_ = b.initFilesRegistry()
	_ = b.initEventsFile()

	t.Cleanup(b.Shutdown)

	_, _ = b.write.file.Write([]byte{1, 2, 3})

	if b.calculateWriteFileSize() != 3 {
		t.Errorf("calculateWriteFileSize failed")
	}
}
