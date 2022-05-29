package binarylog

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	path := t.TempDir()
	binlog, _ := New(path, os.Stderr)

	data := []byte{0, 0, 0}
	_, _ = binlog.Log(data)
	_, _ = binlog.Flush()

	binlog.Shutdown()

	expectedFileSize := binlog.eventsFileSize

	binlog, _ = New(path, os.Stderr)

	if expectedFileSize != binlog.eventsFileSize {
		t.Errorf("eventsFileSize check was failed")
	}

	t.Cleanup(binlog.Shutdown)
}

func Test_binaryLogger_LogCheckRotate(t *testing.T) {
	binlog, _ := New(t.TempDir(), os.Stderr)
	binlog.SetLogFileMaxSize(1)
	t.Cleanup(binlog.Shutdown)

	_, _ = binlog.Log([]byte("some data"))

	if binlog.calculateLogFileSize() != 0 {
		t.Errorf("LogCheckRotate expect create new log file, its must be empty after rotate")
		return
	}

	if len(binlog.buf.Bytes()) > 0 {
		t.Errorf("LogCheckRotate expect flush before log rotate")
	}

	expectedMap := logFilesMap{1: "events.1", 2: "events.2"}
	isEqual := reflect.DeepEqual(binlog.eventsFilesMap, expectedMap)

	if !isEqual {
		t.Errorf("LogCheckRotate not equal expected eventsFilesMap (%v), got %v", expectedMap, binlog.eventsFilesMap)
	}
}

func Test_binaryLogger_autoFlushCount(t *testing.T) {
	binlog, _ := New(t.TempDir(), os.Stderr)
	binlog.SetAutoFlushCount(1)
	t.Cleanup(binlog.Shutdown)

	_, _ = binlog.Log([]byte{0})

	if binlog.calculateLogFileSize() == 0 {
		t.Errorf("autoFlushCount failed, expected to flush")
	}
}

func Test_binaryLogger_autoFlushCountFailedFlush(t *testing.T) {
	binlog := eventStorage{
		buf:         new(bytes.Buffer),
		fileMaxSize: MB,
	}

	binlog.SetAutoFlushCount(1)
	t.Cleanup(binlog.Shutdown)

	if _, err := binlog.Log([]byte{0}); err == nil {
		t.Errorf("autoFlushCountFailedFlush expected error for flush without eventsFile, got nil")
	}
}

func Test_binaryLogger_LogFailedRotateFlush(t *testing.T) {
	binlog := eventStorage{
		buf:         new(bytes.Buffer),
		fileMaxSize: 1,
	}

	t.Cleanup(binlog.Shutdown)

	_, err := binlog.Log([]byte{0})

	if err == nil {
		t.Errorf("LogFailedRotateFlush expected error for flush without eventsFile, got nil")
	}
}

func Test_binaryLogger_autoFlushCountSetterGetter(t *testing.T) {
	binlog := eventStorage{}
	binlog.SetAutoFlushCount(7)

	if binlog.GetAutoFlushCount() != 7 {
		t.Errorf("SetAutoFlushCount failed")
	}
}

func Test_binaryLogger_SetAutoFlushTimeAlreadySet(t *testing.T) {
	binlog, _ := New(t.TempDir(), os.Stderr)
	_ = binlog.SetAutoFlushTime(time.Millisecond)
	t.Cleanup(binlog.Shutdown)

	if err := binlog.SetAutoFlushTime(time.Millisecond); err != ErrAutoFlushTimeAlreadySet {
		t.Errorf("SetAutoFlushTimeAlreadySet failed, expected error")
	}
}
func Test_binaryLogger_SetAutoFlushFailed(t *testing.T) {
	errBuffer := new(bytes.Buffer)
	binlog, _ := New(t.TempDir(), errBuffer)
	t.Cleanup(binlog.Shutdown)

	if err := binlog.SetAutoFlushTime(time.Millisecond); err != nil {
		t.Errorf("SetAutoFlushFailed failed, err: " + err.Error())
		return
	}

	_ = binlog.eventsFile.Close()
	binlog.eventsFile = nil
	_, _ = binlog.Log([]byte{0})
	time.Sleep(time.Millisecond * 100)

	if len(errBuffer.Bytes()) == 0 {
		t.Errorf("SetAutoFlushTime expected or to flush nil file, got nil")
	}
}

func Test_binaryLogger_SetAutoFlushTimeWrongPeriod(t *testing.T) {
	binlog, _ := New(t.TempDir(), os.Stderr)
	t.Cleanup(binlog.Shutdown)

	if err := binlog.SetAutoFlushTime(-time.Millisecond); err != ErrAutoFlushTimeTooLow {
		t.Errorf("SetAutoFlushTimeAlreadySet failed, expected error")
	}
}

func Test_binaryLogger_SetAutoFlushTime(t *testing.T) {
	binlog, _ := New(t.TempDir(), os.Stderr)
	t.Cleanup(binlog.Shutdown)

	if err := binlog.SetAutoFlushTime(time.Millisecond); err != nil {
		t.Errorf("SetAutoFlushTime failed, err: " + err.Error())
		return
	}

	data := []byte("s")
	_, _ = binlog.Log(data)
	time.Sleep(time.Millisecond * 100)

	// todo - Repair

	// raw, _ := binlog.Read(0, int64(hex.EncodedLen(len(data))), 0)
	// decoded, _ := binlog.Decode(raw)
	//
	// if string(decoded) != string(data) {
	// 	t.Errorf("SetAutoFlushTime failed, fetched data is incorrect")
	// }
}

func BenchmarkEventStorage_ReadEvents(b *testing.B) {
	storage, _ := New(b.TempDir(), os.Stderr)
	benchmarksFillBinlog(storage, b)

	b.Cleanup(storage.Shutdown)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		storage.ReadEvents(1, 0)
	}
}

func BenchmarkLog(b *testing.B) {
	binlog := benchmarksInitBinlog(b)
	raw := []byte("asdf asdf asdf asdf asdf")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = binlog.Log(raw)
	}

	b.StopTimer()
	_, _ = binlog.Flush()
}

func benchmarksFillBinlog(binlog *eventStorage, b *testing.B) {
	raw := []byte("some data for tests ")

	for i := 0; i < 100000; i++ { // ~20 MB of data
		_, _ = binlog.Log(raw)
	}

	_, _ = binlog.Flush()
	b.ResetTimer()
}

func benchmarksInitBinlog(b *testing.B) *eventStorage {
	_, base, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(base)

	binlogPath := basepath + string(os.PathSeparator) + "testdata"
	_ = os.MkdirAll(binlogPath, 0755)

	binlog, _ := New(binlogPath, os.Stderr)
	binlog.SetLogFileMaxSize(1000 * MB)

	b.Cleanup(binlog.Shutdown)

	return binlog
}
