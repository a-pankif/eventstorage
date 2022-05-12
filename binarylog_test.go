package binarylog

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
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

	expectedLastLineBytesUsed := len(data) + len(RowDelimiter)
	expectedFileSize := binlog.logFileSize

	binlog, _ = New(path, os.Stderr)

	if expectedLastLineBytesUsed != binlog.lastLineBytesCount {
		t.Errorf("lastLineBytesCount check was failed")
	}

	if expectedFileSize != binlog.logFileSize {
		t.Errorf("logFileSize check was failed")
	}

	t.Cleanup(binlog.Shutdown)
}

func Test_binaryLogger_autoFlushCountSetterGetter(t *testing.T) {
	binlog := binaryLogger{}
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

	_ = binlog.logFile.Close()
	binlog.logFile = nil
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

	raw, _ := binlog.Read(0, int64(hex.EncodedLen(len(data))), 0)
	decoded, _ := binlog.Decode(raw)

	if string(decoded) != string(data) {
		t.Errorf("SetAutoFlushTime failed, fetched data is incorrect")
	}
}

func Test_binaryLogger_ReadToSeekFailed(t *testing.T) {
	binlog, _ := New(t.TempDir(), os.Stderr)
	t.Cleanup(binlog.Shutdown)
	buf := make([]byte, 0)

	if err := binlog.ReadTo(&buf, -5, 0); err == nil {
		t.Errorf("ReadToSeekFailed expected error, got nil")
	}
}

func Test_binaryLogger_ReadFailed(t *testing.T) {
	binlog, _ := New(t.TempDir(), os.Stderr)
	t.Cleanup(binlog.Shutdown)

	if _, err := binlog.Read(0, 10, 0); err == nil {
		t.Errorf("ReadToFailed expected error, got nil")
	}
}

func Test_binaryLogger_ReadToFailed(t *testing.T) {
	binlog, _ := New(t.TempDir(), os.Stderr)
	t.Cleanup(binlog.Shutdown)
	buf := make([]byte, 10)

	if err := binlog.ReadTo(&buf, 0, 0); err == nil {
		t.Errorf("ReadToFailed expected error, got nil")
	}
}

func Test_binaryLogger_DecodeFailed(t *testing.T) {
	binlog := binaryLogger{}
	_, err := binlog.Decode([]byte("7373~7373"))

	if err == nil {
		t.Errorf("DecodeFailed expected error, got nil")
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

func BenchmarkReadTo(b *testing.B) {
	buffer := make([]byte, 1000000)
	binlog := benchmarksInitBinlog(b)
	benchmarksFillBinlog(binlog, b)

	for i := 0; i < b.N; i++ {
		_ = binlog.ReadTo(&buffer, 0, 0)
	}
}

func BenchmarkRead(b *testing.B) {
	binlog := benchmarksInitBinlog(b)
	benchmarksFillBinlog(binlog, b)

	for i := 0; i < b.N; i++ {
		_, _ = binlog.Read(0, 1000000, 0)
	}
}

func BenchmarkDecode(b *testing.B) {
	binlog := benchmarksInitBinlog(b)
	benchmarksFillBinlog(binlog, b)
	data, _ := binlog.Read(0, 1000000, 0)

	for i := 0; i < b.N; i++ {
		_, _ = binlog.Decode(data)
	}
}

func benchmarksFillBinlog(binlog *binaryLogger, b *testing.B) {
	raw := []byte("some data for tests ")

	for i := 0; i < 10000; i++ { // ~2 MB of data
		_, _ = binlog.Log(raw)
	}

	_, _ = binlog.Flush()
	b.ResetTimer()
}

func benchmarksInitBinlog(b *testing.B) *binaryLogger {
	_, base, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(base)

	binlogPath := basepath + string(os.PathSeparator) + "testdata"
	_ = os.MkdirAll(binlogPath, 0755)

	binlog, _ := New(binlogPath, os.Stderr)
	binlog.SetLogFileMaxSize(1000 * MB)

	b.Cleanup(binlog.Shutdown)

	return binlog
}
