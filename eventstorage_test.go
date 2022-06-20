package eventstorage

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	path := t.TempDir()
	storage, _ := New(path)

	_, _ = storage.Write([]byte{0, 0, 0})
	_, _ = storage.Flush()

	storage.Shutdown()

	expectedFileSize := storage.write.fileSize
	storage, _ = New(path)

	if expectedFileSize != storage.write.fileSize {
		t.Errorf("fileSize check was failed")
	}

	t.Cleanup(storage.Shutdown)
}

func Test_eventStorage_WriteCheckRotate(t *testing.T) {
	storage, _ := New(t.TempDir())
	storage.SetWriteFileMaxSize(1)
	t.Cleanup(storage.Shutdown)

	_, _ = storage.Write([]byte("some data"))
	_, _ = storage.Write([]byte("some data"))
	_, _ = storage.Write([]byte("some data"))
	_, _ = storage.Write([]byte("some data"))
	_, _ = storage.Write([]byte("some data"))
	_, _ = storage.Write([]byte("some data"))

	if storage.calculateWriteFileSize() != 0 {
		t.Errorf("WriteCheckRotate expect create new events file, it must be empty after rotate")
		return
	}

	if len(storage.write.buf.Bytes()) > 0 {
		t.Errorf("WriteCheckRotate expect flush before events file rotate")
		return
	}

	if len(storage.read.readableFiles) != 7 {
		t.Errorf("WriteCheckRotate expect 7 events files, got %v", len(storage.read.readableFiles))
		return
	}

	for i, file := range storage.read.readableFiles {
		info, _ := file.Stat()
		expectedName := fmt.Sprintf(eventsFileNameTemplate, i)
		if info.Name() != expectedName {
			t.Errorf("WriteCheckRotate not equal expected events file name (%v), got %v", expectedName, info.Name())
		}
	}
}

func Test_eventStorage_Read(t *testing.T) {
	storage, _ := New(t.TempDir())
	storage.SetWriteFileMaxSize(20)
	storage.SetAutoFlushCount(1)
	t.Cleanup(storage.Shutdown)

	const dataPrefix = "some data"
	const iterCount = 300
	const offset = 10

	for i := 0; i <= iterCount; i++ {
		_, _ = storage.Write([]byte(dataPrefix + strconv.Itoa(i)))
	}

	events := storage.Read(iterCount-offset, offset)
	for i, event := range events {
		if event != dataPrefix+strconv.Itoa(offset+i) {
			t.Errorf("Read failed, incorrect data.")
			return
		}
	}
}

func Test_eventStorage_autoFlushCount(t *testing.T) {
	storage, _ := New(t.TempDir())
	storage.SetAutoFlushCount(1)
	t.Cleanup(storage.Shutdown)

	_, _ = storage.Write([]byte{0})

	if storage.calculateWriteFileSize() == 0 {
		t.Errorf("autoFlushCount failed, expected to flush")
	}
}

func Test_eventStorage_autoFlushCountFailedFlush(t *testing.T) {
	storage := eventStorage{
		write: &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:  &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}

	storage.SetAutoFlushCount(1)
	t.Cleanup(storage.Shutdown)

	if _, err := storage.Write([]byte{0}); err == nil {
		t.Errorf("autoFlushCountFailedFlush expected error for flush without file, got nil")
	}
}

func Test_eventStorage_WriteFailedRotateFlush(t *testing.T) {
	storage := eventStorage{
		write: &write{buf: new(bytes.Buffer), fileMaxSize: 1},
		read:  &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}

	t.Cleanup(storage.Shutdown)

	_, err := storage.Write([]byte{0})

	if err == nil {
		t.Errorf("WriteFailedRotateFlush expected error for flush without file, got nil")
	}
}

func Test_eventStorage_autoFlushCountSetterGetter(t *testing.T) {
	storage := eventStorage{
		write: &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:  &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
	}
	storage.SetAutoFlushCount(7)

	if storage.GetAutoFlushCount() != 7 {
		t.Errorf("SetAutoFlushCount failed")
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
	_, _ = storage.Write(data)
	time.Sleep(time.Millisecond * 100)

	events := storage.Read(1, 0)

	if len(events) == 0 {
		t.Errorf("SetAutoFlushTime failed, fetched data is incorrect")
		return
	}

	isEqual := reflect.DeepEqual(events[0], string(data))

	if !isEqual {
		t.Errorf("SetAutoFlushTime failed, fetched data is incorrect")
	}
}

func BenchmarkWriteChar(b *testing.B) {
	storage := benchmarksInitStorage(b)
	raw := []byte("s")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = storage.Write(raw)
	}

	b.StopTimer()
	_, _ = storage.Flush()
}

func BenchmarkEventStorage_ReadChar(b *testing.B) {
	storage, _ := New(b.TempDir())
	benchmarksFillstorage(storage, b)

	b.Cleanup(storage.Shutdown)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		storage.Read(1, 0)
	}
}

func BenchmarkEventStorage_CharReadTo(b *testing.B) {
	storage, _ := New(b.TempDir())
	readTo := make([]string, 1)
	benchmarksFillstorage(storage, b)

	b.Cleanup(storage.Shutdown)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		storage.ReadTo(1, 0, readTo)
	}
}

func BenchmarkEventStorage_CharReadToOffset10000(b *testing.B) {
	storage, _ := New(b.TempDir())
	readTo := make([]string, 1)
	benchmarksFillstorage(storage, b)
	b.Cleanup(storage.Shutdown)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		storage.ReadTo(1, 10000, readTo)
	}
}

func benchmarksFillstorage(storage *eventStorage, b *testing.B) {
	raw := []byte("s")

	for i := 0; i < 300000; i++ {
		_, _ = storage.Write(raw)
	}

	_, _ = storage.Flush()
	b.ResetTimer()
}

func benchmarksInitStorage(b *testing.B) *eventStorage {
	storage, _ := New(b.TempDir())
	storage.SetWriteFileMaxSize(1000 * MB)
	b.Cleanup(storage.Shutdown)

	return storage
}
