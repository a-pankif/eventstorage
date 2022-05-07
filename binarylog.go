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

func New(logFile *os.File, errWriter io.Writer, logWriter io.Writer) *blogger {
	b := &blogger{
		buf:       new(bytes.Buffer),
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
	b.lastLineUsed = len(res)

	return b
}

func (b *blogger) Log(data []byte) {
	dist := new(bytes.Buffer)
	encodeBuf := make([]byte, 2) // 2 bytes for HEX data, 1 byte takes 2 bytes in HEX

	for i := range data {
		hex.Encode(encodeBuf, data[i:i+1])
		dist.Write(encodeBuf)

		b.lastLineUsed++

		if b.lastLineUsed >= 16 {
			dist.Write([]byte{'\n'})
			b.lastLineUsed = 0
		} else if b.lastLineUsed%2 == 0 {
			dist.Write([]byte{' '}) // Group by 2 bytes.
		}
	}

	b.bufLock.Lock()
	b.buf.Write(dist.Bytes())
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

func (b *blogger) Flush() (count int) {
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

func (b *blogger) SetAutoFlushCount(count int) {
	b.autoFlushCount = count
}

func (b *blogger) GetAutoFlushCount() int {
	return b.autoFlushCount
}

func (b *blogger) SetAutoFlushTime(period time.Duration) error {
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
