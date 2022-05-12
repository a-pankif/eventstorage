package binarylog

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func BenchmarkLog(b *testing.B) {
	binlog := benchmarksInitBinlog(b)
	raw := []byte("asdf asdf asdf asdf asdf")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		binlog.Log(raw)
	}

	b.StopTimer()
	binlog.Flush()
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

func BenchmarkDecodeLen(b *testing.B) {
	binlog := benchmarksInitBinlog(b)
	benchmarksFillBinlog(binlog, b)
	data, _ := binlog.Read(0, 1000000, 0)

	for i := 0; i < b.N; i++ {
		_ = binlog.DecodeLen(data)
	}
}

func BenchmarkDecode(b *testing.B) {
	binlog := benchmarksInitBinlog(b)
	benchmarksFillBinlog(binlog, b)
	data, _ := binlog.Read(0, 1000000, 0)

	for i := 0; i < b.N; i++ {
		_ = binlog.Decode(data)
	}
}

func benchmarksFillBinlog(binlog *binaryLogger, b *testing.B) {
	raw := []byte("some data for tests ")

	for i := 0; i < 10000; i++ { // ~2 MB of data
		binlog.Log(raw)
	}

	binlog.Flush()
	b.ResetTimer()
}

func benchmarksInitBinlog(b *testing.B) *binaryLogger {
	_, base, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(base)

	binlogPath := basepath + string(os.PathSeparator) + "testdata"
	_ = os.MkdirAll(binlogPath, 0755)

	binlog, _ := New(binlogPath, os.Stderr, os.Stdout)
	binlog.SetLogFileMaxSize(1000 * MB)

	b.Cleanup(func() {
		_ = binlog.logFile.Close()
		_ = os.Remove(binlogPath)
	})

	return binlog
}
