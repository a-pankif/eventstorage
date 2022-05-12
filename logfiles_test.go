package binarylog

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"testing"
)

func Test_binaryLogger_initRegistryFile(t *testing.T) {
	b := &binaryLogger{
		basePath:    t.TempDir(),
		logFilesMap: make(logFilesMap),
	}

	if err := b.initRegistryFile(); err != nil {
		t.Errorf("logFilesRegistry not inited")
		return
	}

	t.Cleanup(func() {
		_ = b.logFilesRegistry.Close()
	})
}

func Test_binaryLogger_initRegistryFileFailed(t *testing.T) {
	b := &binaryLogger{
		basePath:    string([]byte{0}),
		logFilesMap: make(logFilesMap),
	}

	if err := b.initRegistryFile(); err == nil {
		t.Cleanup(func() {
			_ = b.logFilesRegistry.Close()
		})
		t.Errorf("logFilesRegistry inited, but expected fail")
	}
}

func Test_binaryLogger_appendInRegistryFile(t *testing.T) {
	b := &binaryLogger{
		basePath:    t.TempDir(),
		logFilesMap: make(logFilesMap),
		errWriter:   os.Stderr,
	}

	_ = b.initRegistryFile()

	b.logFilesCount++
	file1, _ := b.openLogFile(b.logFilesCount, true)
	b.logFilesCount++
	file2, _ := b.openLogFile(b.logFilesCount, true)

	_ = file1.Close()
	_ = file2.Close()
	_ = b.logFilesRegistry.Close()

	b.logFilesCount = 0
	b.logFilesMap = make(logFilesMap)
	_ = b.initRegistryFile()

	expectedMap := logFilesMap{1: "binlog.1", 2: "binlog.2"}
	isEqual := reflect.DeepEqual(b.logFilesMap, expectedMap)

	if !isEqual {
		t.Errorf("logFilesMap not equal expected")
	}

	t.Cleanup(func() {
		_ = b.logFilesRegistry.Close()
	})
}

func Test_binaryLogger_initLogFileWithoutRegistry(t *testing.T) {
	b := &binaryLogger{}

	if err := b.initLogFile(); err == nil {
		t.Errorf("initLogFile expected failed without registry")
	}
}

func Test_binaryLogger_initLogFile(t *testing.T) {
	b := &binaryLogger{
		basePath:    t.TempDir(),
		logFilesMap: make(logFilesMap),
		errWriter:   os.Stderr,
	}
	_ = b.initRegistryFile()

	if err := b.initLogFile(); err != nil {
		t.Errorf("initLogFile failed")
		return
	}

	t.Cleanup(func() {
		_ = b.logFilesRegistry.Close()
		_ = b.CloseLogFile()
	})
}

func Test_binaryLogger_initLogFileFailed(t *testing.T) {
	b := &binaryLogger{
		basePath:    t.TempDir(),
		logFilesMap: make(logFilesMap),
		errWriter:   os.Stderr,
	}

	_ = b.initRegistryFile()

	t.Cleanup(func() {
		_ = b.logFilesRegistry.Close()
	})

	b.basePath = string([]byte{0})

	if err := b.initLogFile(); err == nil {
		t.Errorf("initLogFile expect failed")
	}
}

func Test_binaryLogger_rotateLogFileFailedCloseOld(t *testing.T) {
	b := &binaryLogger{}

	if err := b.rotateLogFile(); err == nil {
		t.Errorf("rotateLogFileFailedCloseOld expect failed")
	}
}

func Test_binaryLogger_rotateLogFile(t *testing.T) {
	b := &binaryLogger{
		basePath:    t.TempDir(),
		logFilesMap: make(logFilesMap),
	}

	_ = b.initRegistryFile()
	_ = b.initLogFile()

	t.Cleanup(func() {
		_ = b.logFilesRegistry.Close()
		_ = b.CloseLogFile()
	})

	if err := b.rotateLogFile(); err != nil {
		t.Errorf("rotateLogFile failed")
		return
	}

	expectedMap := logFilesMap{1: "binlog.1", 2: "binlog.2"}
	isEqual := reflect.DeepEqual(b.logFilesMap, expectedMap)

	if !isEqual {
		t.Errorf("rotateLogFile logFilesMap not equal expected")
	}
}

func Test_binaryLogger_openLogFileFailedAppend(t *testing.T) {
	b := &binaryLogger{}

	if _, err := b.openLogFile(1, true); err == nil {
		t.Errorf("openLogFileFailedAppend expect failed")
	}
}

func Test_binaryLogger_logErrorString(t *testing.T) {
	buf := new(bytes.Buffer)
	b := &binaryLogger{
		errWriter: buf,
	}

	b.logErrorString("error")

	if len(buf.Bytes()) != 6 { // Six symbol is 0a (\r)
		t.Errorf("logErrorString wrong error data ")
	}
}

func Test_binaryLogger_logError(t *testing.T) {
	buf := new(bytes.Buffer)
	b := &binaryLogger{
		errWriter: buf,
	}

	b.logError(errors.New("error"))

	if len(buf.Bytes()) != 6 { // Six symbol is 0a (\r)
		t.Errorf("logErrorString wrong error data ")
	}
}

func Test_binaryLogger_SetLogFileSize(t *testing.T) {
	b := &binaryLogger{}
	b.SetLogFileMaxSize(100)

	if b.logFileMaxSize != 100 {
		t.Errorf("SetLogFileMaxSize failed")
	}
}

func Test_binaryLogger_calculateLogFileSize(t *testing.T) {
	b := &binaryLogger{
		basePath:    t.TempDir(),
		logFilesMap: make(logFilesMap),
	}

	_ = b.initRegistryFile()
	_ = b.initLogFile()

	t.Cleanup(func() {
		_ = b.logFilesRegistry.Close()
		_ = b.CloseLogFile()
	})

	_, _ = b.logFile.Write([]byte{1, 2, 3})

	if b.calculateLogFileSize() != 3 {
		t.Errorf("calculateLogFileSize failed")
	}
}
