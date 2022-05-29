package binarylog

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"testing"
)

func Test_binaryLogger_initRegistryFile(t *testing.T) {
	b := &eventStorage{
		basePath:       t.TempDir(),
		eventsFilesMap: make(logFilesMap),
	}

	if err := b.initRegistryFile(); err != nil {
		t.Errorf("eventsFilesRegistry not inited")
		return
	}

	t.Cleanup(func() {
		_ = b.eventsFilesRegistry.Close()
	})
}

func Test_binaryLogger_initRegistryFileFailed(t *testing.T) {
	b := &eventStorage{
		basePath:       string([]byte{0}),
		eventsFilesMap: make(logFilesMap),
	}

	if err := b.initRegistryFile(); err == nil {
		t.Cleanup(func() {
			_ = b.eventsFilesRegistry.Close()
		})
		t.Errorf("eventsFilesRegistry inited, but expected fail")
	}
}

func Test_binaryLogger_appendInRegistryFile(t *testing.T) {
	b := &eventStorage{
		basePath:       t.TempDir(),
		eventsFilesMap: make(logFilesMap),
		errWriter:      os.Stderr,
	}

	_ = b.initRegistryFile()

	b.filesCount++
	file1, _ := b.openEventsFile(b.filesCount, true)
	b.filesCount++
	file2, _ := b.openEventsFile(b.filesCount, true)

	_ = file1.Close()
	_ = file2.Close()
	_ = b.eventsFilesRegistry.Close()

	b.filesCount = 0
	b.eventsFilesMap = make(logFilesMap)
	_ = b.initRegistryFile()

	expectedMap := logFilesMap{1: "binlog.1", 2: "binlog.2"}
	isEqual := reflect.DeepEqual(b.eventsFilesMap, expectedMap)

	if !isEqual {
		t.Errorf("eventsFilesMap not equal expected")
	}

	t.Cleanup(func() {
		_ = b.eventsFilesRegistry.Close()
	})
}

func Test_binaryLogger_initLogFileWithoutRegistry(t *testing.T) {
	b := &eventStorage{}

	if err := b.initEventsFile(); err == nil {
		t.Errorf("initEventsFile expected failed without registry")
	}
}

func Test_binaryLogger_initLogFile(t *testing.T) {
	b := &eventStorage{
		basePath:       t.TempDir(),
		eventsFilesMap: make(logFilesMap),
		errWriter:      os.Stderr,
	}
	_ = b.initRegistryFile()

	if err := b.initEventsFile(); err != nil {
		t.Errorf("initEventsFile failed")
		return
	}

	t.Cleanup(func() {
		_ = b.eventsFilesRegistry.Close()
		_ = b.CloseLogFile()
	})
}

func Test_binaryLogger_initLogFileFailed(t *testing.T) {
	b := &eventStorage{
		basePath:       t.TempDir(),
		eventsFilesMap: make(logFilesMap),
		errWriter:      os.Stderr,
	}

	_ = b.initRegistryFile()

	t.Cleanup(func() {
		_ = b.eventsFilesRegistry.Close()
	})

	b.basePath = string([]byte{0})

	if err := b.initEventsFile(); err == nil {
		t.Errorf("initEventsFile expect failed")
	}
}

func Test_binaryLogger_rotateLogFileFailedCloseOld(t *testing.T) {
	b := &eventStorage{}

	if err := b.rotateEventsFile(); err == nil {
		t.Errorf("rotateLogFileFailedCloseOld expect failed")
	}
}

func Test_binaryLogger_rotateLogFile(t *testing.T) {
	b := &eventStorage{
		basePath:       t.TempDir(),
		eventsFilesMap: make(logFilesMap),
	}

	_ = b.initRegistryFile()
	_ = b.initEventsFile()

	t.Cleanup(func() {
		_ = b.eventsFilesRegistry.Close()
		_ = b.CloseLogFile()
	})

	if err := b.rotateEventsFile(); err != nil {
		t.Errorf("rotateEventsFile failed")
		return
	}

	expectedMap := logFilesMap{1: "binlog.1", 2: "binlog.2"}
	isEqual := reflect.DeepEqual(b.eventsFilesMap, expectedMap)

	if !isEqual {
		t.Errorf("rotateEventsFile eventsFilesMap not equal expected")
	}
}

func Test_binaryLogger_openLogFileFailedAppend(t *testing.T) {
	b := &eventStorage{}

	if _, err := b.openEventsFile(1, true); err == nil {
		t.Errorf("openLogFileFailedAppend expect failed")
	}
}

func Test_binaryLogger_logErrorString(t *testing.T) {
	buf := new(bytes.Buffer)
	b := &eventStorage{
		errWriter: buf,
	}

	b.logErrorString("error")

	if len(buf.Bytes()) != 6 { // Six symbol is 0a (\r)
		t.Errorf("logErrorString wrong error data ")
	}
}

func Test_binaryLogger_logError(t *testing.T) {
	buf := new(bytes.Buffer)
	b := &eventStorage{
		errWriter: buf,
	}

	b.logError(errors.New("error"))

	if len(buf.Bytes()) != 6 { // Six symbol is 0a (\r)
		t.Errorf("logErrorString wrong error data ")
	}
}

func Test_binaryLogger_SetLogFileSize(t *testing.T) {
	b := &eventStorage{}
	b.SetLogFileMaxSize(100)

	if b.fileMaxSize != 100 {
		t.Errorf("SetLogFileMaxSize failed")
	}
}

func Test_binaryLogger_calculateLogFileSize(t *testing.T) {
	b := &eventStorage{
		basePath:       t.TempDir(),
		eventsFilesMap: make(logFilesMap),
	}

	_ = b.initRegistryFile()
	_ = b.initEventsFile()

	t.Cleanup(func() {
		_ = b.eventsFilesRegistry.Close()
		_ = b.CloseLogFile()
	})

	_, _ = b.eventsFile.Write([]byte{1, 2, 3})

	if b.calculateLogFileSize() != 3 {
		t.Errorf("calculateLogFileSize failed")
	}
}
