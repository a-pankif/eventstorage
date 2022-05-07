package binarylog

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// func BenchmarkLog(b *testing.B) {
// 	binlog := testsInitBinlog(b)
// 	raw := []byte("asdf asdf asdf asdf asdf")
//
// 	b.ResetTimer()
//
// 	for i := 0; i < b.N; i++ {
// 		binlog.Log(raw)
// 	}
//
// 	b.StopTimer()
// 	binlog.Flush()
// }

func BenchmarkReadTo(b *testing.B) {
	binlog := testsInitBinlog(b)
	raw := []byte("some data for tests ")
	buffer := make([]byte, 1000000)

	for i := 0; i < 10000; i++ { // ~2 MB of data
		binlog.Log(raw)
	}

	binlog.Flush()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = binlog.ReadTo(&buffer, 0, 0)
	}
}

func BenchmarkRead(b *testing.B) {
	binlog := testsInitBinlog(b)
	raw := []byte("some data for tests ")

	for i := 0; i < 10000; i++ { // ~2 MB of data
		binlog.Log(raw)
	}

	binlog.Flush()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = binlog.Read(0, 1000000, 0)
	}
}

func testsInitBinlog(b *testing.B) *binaryLogger {
	_, base, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(base)

	binlogPath := basepath + string(os.PathSeparator) + "testdata"
	binlogFullPath := binlogPath + string(os.PathSeparator) + "test-binlog.1"
	_ = os.MkdirAll(binlogPath, 0755)

	binlogFile, _ := os.OpenFile(binlogFullPath, os.O_CREATE|os.O_APPEND, 0644)
	binlog := New(binlogFile, os.Stderr, os.Stdout)

	b.Cleanup(func() {
		_ = binlog.logFile.Close()
		// _ = os.Remove(binlogFullPath)
	})

	return binlog
}
