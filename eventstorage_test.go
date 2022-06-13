package eventstorage

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	path := t.TempDir()
	storage, _ := New(path)

	data := []byte{0, 0, 0}
	_, _ = storage.Log(data)
	_, _ = storage.Flush()

	storage.Shutdown()

	expectedFileSize := storage.write.fileSize
	storage, _ = New(path)

	if expectedFileSize != storage.write.fileSize {
		t.Errorf("fileSize check was failed")
	}

	t.Cleanup(storage.Shutdown)
}

func Test_eventStorage_LogCheckRotate(t *testing.T) {
	storage, _ := New(t.TempDir())
	storage.SetLogFileMaxSize(1)
	t.Cleanup(storage.Shutdown)

	_, _ = storage.Log([]byte("some data"))
	_, _ = storage.Log([]byte("some data"))
	_, _ = storage.Log([]byte("some data"))
	_, _ = storage.Log([]byte("some data"))
	_, _ = storage.Log([]byte("some data"))
	_, _ = storage.Log([]byte("some data"))

	if storage.calculateLogFileSize() != 0 {
		t.Errorf("LogCheckRotate expect create new log file, its must be empty after rotate")
		return
	}

	if len(storage.write.buf.Bytes()) > 0 {
		t.Errorf("LogCheckRotate expect flush before log rotate")
		return
	}

	if len(storage.read.readableFiles) != 7 {
		t.Errorf("LogCheckRotate expect 7 events files, got %v", len(storage.read.readableFiles))
		return
	}

	for i, file := range storage.read.readableFiles {
		info, _ := file.Stat()
		expectedName := fmt.Sprintf(eventsFileNameTemplate, i)
		if info.Name() != expectedName {
			t.Errorf("LogCheckRotate not equal expected events file name (%v), got %v", expectedName, info.Name())
		}
	}
}

func Test_eventStorage_autoFlushCount(t *testing.T) {
	storage, _ := New(t.TempDir())
	storage.setAutoFlushCount(1)
	t.Cleanup(storage.Shutdown)

	_, _ = storage.Log([]byte{0})

	if storage.calculateLogFileSize() == 0 {
		t.Errorf("autoFlushCount failed, expected to flush")
	}
}

func Test_eventStorage_autoFlushCountFailedFlush(t *testing.T) {
	storage := eventStorage{
		write: &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:  &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}

	storage.setAutoFlushCount(1)
	t.Cleanup(storage.Shutdown)

	if _, err := storage.Log([]byte{0}); err == nil {
		t.Errorf("autoFlushCountFailedFlush expected error for flush without file, got nil")
	}
}

func Test_eventStorage_LogFailedRotateFlush(t *testing.T) {
	storage := eventStorage{
		write: &write{buf: new(bytes.Buffer), fileMaxSize: 1},
		read:  &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}

	t.Cleanup(storage.Shutdown)

	_, err := storage.Log([]byte{0})

	if err == nil {
		t.Errorf("LogFailedRotateFlush expected error for flush without file, got nil")
	}
}

func Test_eventStorage_autoFlushCountSetterGetter(t *testing.T) {
	storage := eventStorage{
		write: &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:  &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}
	storage.setAutoFlushCount(7)

	if storage.GetAutoFlushCount() != 7 {
		t.Errorf("setAutoFlushCount failed")
	}
}

func Test_eventStorage_SetAutoFlushTimeAlreadySet(t *testing.T) {
	storage, _ := New(t.TempDir())
	_ = storage.SetAutoFlushTime(time.Millisecond)
	t.Cleanup(storage.Shutdown)

	if err := storage.SetAutoFlushTime(time.Millisecond); err != ErrAutoFlushTimeAlreadySet {
		t.Errorf("SetAutoFlushTimeAlreadySet failed, expected error")
	}
}

func Test_eventStorage_SetAutoFlushTimeWrongPeriod(t *testing.T) {
	storage, _ := New(t.TempDir())
	t.Cleanup(storage.Shutdown)

	if err := storage.SetAutoFlushTime(-time.Millisecond); err != ErrAutoFlushTimeTooLow {
		t.Errorf("SetAutoFlushTimeAlreadySet failed, expected error")
	}
}

func Test_eventStorage_SetAutoFlushTime(t *testing.T) {
	storage, _ := New(t.TempDir())
	t.Cleanup(storage.Shutdown)

	if err := storage.SetAutoFlushTime(time.Millisecond); err != nil {
		t.Errorf("SetAutoFlushTime failed, err: " + err.Error())
		return
	}

	data := []byte("s")
	_, _ = storage.Log(data)
	time.Sleep(time.Millisecond * 100)

	events := storage.ReadEvents(1, 0)

	isEqual := reflect.DeepEqual(events[0], string(data))

	if !isEqual {
		t.Errorf("SetAutoFlushTime failed, fetched data is incorrect")
	}
}

func BenchmarkEventStorage_ReadEvents(b *testing.B) {
	storage, _ := New(b.TempDir())
	benchmarksFillstorage(storage, b)

	b.Cleanup(storage.Shutdown)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		storage.ReadEvents(1, 0)
	}
}

func BenchmarkLog(b *testing.B) {
	storage := benchmarksInitStorage(b)
	raw := []byte("asdf asdf asdf asdf asdf")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = storage.Log(raw)
	}

	b.StopTimer()
	_, _ = storage.Flush()
}

func benchmarksFillstorage(storage *eventStorage, b *testing.B) {
	raw := []byte("some data for tests ")

	for i := 0; i < 100000; i++ { // ~20 MB of data
		_, _ = storage.Log(raw)
	}

	_, _ = storage.Flush()
	b.ResetTimer()
}

func benchmarksInitStorage(b *testing.B) *eventStorage {
	storage, _ := New(b.TempDir())
	storage.SetLogFileMaxSize(1000 * MB)
	b.Cleanup(storage.Shutdown)

	return storage
}
