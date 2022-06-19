package eventstorage

import (
	"bytes"
	"strings"
	"testing"
)

func Test_eventStorage_initRegistryFile(t *testing.T) {
	s := &eventStorage{
		basePath: t.TempDir(),
		write:    &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:     &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}

	t.Cleanup(s.Shutdown)

	if err := s.initFilesRegistry(); err != nil {
		t.Errorf("filesRegistry not inited")
		return
	}
}

func Test_eventStorage_initRegistryFileFailed(t *testing.T) {
	s := &eventStorage{
		basePath: string([]byte{0}),
		write:    &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:     &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}
	t.Cleanup(s.Shutdown)

	if err := s.initFilesRegistry(); err == nil {
		t.Errorf("filesRegistry inited, but expected fail")
	}
}

func Test_eventStorage_appendInRegistryFile(t *testing.T) {
	s := &eventStorage{
		basePath: t.TempDir(),
		write:    &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:     &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}
	t.Cleanup(s.Shutdown)
	_ = s.initFilesRegistry()

	file1, _ := s.openEventsFile(1, true)
	file2, _ := s.openEventsFile(2, true)
	file3, _ := s.openEventsFile(3, true)

	_ = file1.Close()
	_ = file2.Close()
	_ = file3.Close()
	_ = s.filesRegistry.Close()

	if len(s.read.readableFiles) != 3 {
		t.Errorf("appendInRegistryFile has wrong count")
		return
	}

	for i, file := range s.read.readableFiles {
		info, _ := file.Stat()

		if s.getFileName(i) != info.Name() {
			t.Errorf("Wrong file name, expected %v, got %v", s.getFileName(i), info.Name())
			return
		}
	}
}

func Test_eventStorage_initLogFileWithoutRegistry(t *testing.T) {
	s := &eventStorage{
		write: &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:  &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}

	if err := s.initEventsFile(); err == nil {
		t.Errorf("initEventsFile expected failed without registry")
	}

	t.Cleanup(s.Shutdown)
}

func Test_eventStorage_initLogFile(t *testing.T) {
	s := &eventStorage{
		basePath: t.TempDir(),
		write:    &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:     &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}
	_ = s.initFilesRegistry()

	if err := s.initEventsFile(); err != nil {
		t.Errorf("initEventsFile failed")
		return
	}

	t.Cleanup(s.Shutdown)
}

func Test_eventStorage_initLogFileFailed(t *testing.T) {
	s := &eventStorage{
		basePath: t.TempDir(),
		write:    &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:     &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}

	_ = s.initFilesRegistry()

	t.Cleanup(s.Shutdown)

	s.basePath = string([]byte{0})

	if err := s.initEventsFile(); err == nil {
		t.Errorf("initEventsFile expect failed")
	}
}

func Test_eventStorage_rotateLogFileFailedCloseOld(t *testing.T) {
	s := &eventStorage{
		write: &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:  &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}

	if err := s.rotateEventsFile(); err == nil {
		t.Errorf("rotateLogFileFailedCloseOld expect failed")
	}

	t.Cleanup(s.Shutdown)
}

func Test_eventStorage_rotateLogFile(t *testing.T) {
	s := &eventStorage{
		basePath: t.TempDir(),
		write:    &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:     &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}

	_ = s.initFilesRegistry()
	_ = s.initEventsFile()

	t.Cleanup(s.Shutdown)

	if err := s.rotateEventsFile(); err != nil {
		t.Errorf("rotateEventsFile failed")
		return
	}

	if len(s.read.readableFiles) != 2 {
		t.Errorf("rotateLogFile has wrong count")
		return
	}

	for i, file := range s.read.readableFiles {
		info, _ := file.Stat()

		if s.getFileName(i) != info.Name() {
			t.Errorf("Wrong file name, expected %v, got %v", s.getFileName(i), info.Name())
			return
		}
	}
}

func Test_eventStorage_openLogFileFailedAppend(t *testing.T) {
	s := &eventStorage{
		write: &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:  &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}
	t.Cleanup(s.Shutdown)

	if _, err := s.openEventsFile(1, true); err == nil {
		t.Errorf("openLogFileFailedAppend expect failed")
	}
}

func Test_eventStorage_SetLogFileSize(t *testing.T) {
	s := &eventStorage{
		write: &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:  &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}
	s.SetWriteFileMaxSize(100)
	t.Cleanup(s.Shutdown)

	if s.write.fileMaxSize != 100 {
		t.Errorf("SetWriteFileMaxSize failed")
	}
}

func Test_eventStorage_calculateLogFileSize(t *testing.T) {
	s, err := New(t.TempDir())

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	s.SetAutoFlushCount(1)
	t.Cleanup(s.Shutdown)

	_, _ = s.Write([]byte{1, 2, 3})

	if s.calculateWriteFileSize() != 4 { // 3 is data length + 1 is line break
		t.Errorf("calculateWriteFileSize failed")
	}
}
