package binarylog

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func New(logFile *os.File, errWriter io.Writer, logWriter io.Writer) *binaryLogger {
	b := &binaryLogger{
		buf:       new(bytes.Buffer),
		encodeBuf: make([]byte, 2),
		logFile:   logFile,
		errWriter: errWriter,
		logWriter: logWriter,
	}

	info, _ := logFile.Stat()
	lineBytesUsed := 0
	lastLine := info.Size() / lineLength
	lineBuffer := make([]byte, lineLength)
	_, _ = logFile.ReadAt(lineBuffer, lastLine*lineLength)

	for _, v := range lineBuffer {
		if v != 0 {
			lineBytesUsed++
		}
	}

	rawLine := strings.NewReplacer(" ", "", "\n", "").Replace(string(lineBuffer))
	res, _ := hex.DecodeString(rawLine)
	b.lastLineBytesCount = len(res)

	return b
}

func (b *binaryLogger) Log(data []byte) {
	b.bufLock.Lock()

	for i := range data {
		hex.Encode(b.encodeBuf, data[i:i+1])
		b.buf.Write(b.encodeBuf)

		b.lastLineBytesCount++

		if b.lastLineBytesCount >= 16 {
			b.buf.WriteByte('\n')
			b.lastLineBytesCount = 0
		} else if b.lastLineBytesCount%2 == 0 {
			b.buf.WriteByte(' ') // Group by 2 bytes.
		}
	}

	b.insertsCount++

	if b.autoFlushCount > 0 {
		if b.insertsCount >= b.autoFlushCount {
			b.bufLock.Unlock()
			b.Flush()
		} else {
			b.bufLock.Unlock()
		}
	} else {
		b.bufLock.Unlock()
	}
}

func (b *binaryLogger) Flush() (count int) {
	b.bufLock.Lock()

	if b.insertsCount > 0 {
		if _, err := b.logFile.Write(b.buf.Bytes()); err != nil {
			_, _ = fmt.Fprint(b.errWriter, err.Error(), "\n")
		} else {
			b.buf.Truncate(0)
			count = b.insertsCount
			b.insertsCount = 0
		}
	}

	b.bufLock.Unlock()
	return
}

func (b *binaryLogger) SetAutoFlushCount(count int) {
	b.autoFlushCount = count
}

func (b *binaryLogger) GetAutoFlushCount() int {
	return b.autoFlushCount
}

func (b *binaryLogger) SetAutoFlushTime(period time.Duration) error {
	if b.autoFlushTime != 0 {
		return ErrAutoFlushTimeAlreadySet
	}

	if period <= 0 {
		return ErrAutoFlushTimeTooLow
	}

	b.autoFlushTime = period

	go func() {
		for range time.Tick(period) {
			count := b.Flush()

			if count > 0 {
				_, _ = fmt.Fprintf(b.logWriter, "Flushed by time: %d.\n", count)
			}
		}
	}()

	return nil
}
