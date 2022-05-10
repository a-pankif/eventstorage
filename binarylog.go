package binarylog

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"
)

func New(basePath string, errWriter io.Writer, logWriter io.Writer) (*binaryLogger, error) {
	b := &binaryLogger{
		basePath:       basePath,
		buf:            new(bytes.Buffer),
		encodeBuf:      make([]byte, 3),
		logFilesCount:  0,
		logFileMaxSize: 100 * MB,
		errWriter:      errWriter,
		logWriter:      logWriter,
	}

	b.initRegistryFile()
	b.initCurrenLogFile()

	if b.currentLogFile == nil {
		return nil, ErrLogNotInited
	}

	b.currenLogFileSize = b.calculateCurrenLogFileSize()

	lineBytesUsed := 0
	lastLine := b.currenLogFileSize / lineLength
	lineBuffer := make([]byte, lineLength)
	_, _ = b.currentLogFile.ReadAt(lineBuffer, lastLine*lineLength)

	for _, v := range lineBuffer {
		if v != 0 {
			lineBytesUsed++
		}
	}

	rawLine := strings.NewReplacer(" ", "", "\n", "").Replace(string(lineBuffer))
	res, _ := hex.DecodeString(rawLine)
	b.lastLineBytesCount = len(res)

	return b, nil
}

func (b *binaryLogger) insertData(data []byte) int64 {
	var dataLen int64 = 0

	for i := range data {
		var l int64 = 2

		hex.Encode(b.encodeBuf, data[i:i+1])
		b.lastLineBytesCount++

		if b.lastLineBytesCount >= 16 {
			b.encodeBuf[2] = '\n'
			b.lastLineBytesCount = 0
			l++
		} else if b.lastLineBytesCount%2 == 0 {
			b.encodeBuf[2] = ' '
			l++
		}

		dataLen += l
		b.buf.Write(b.encodeBuf[:l])
	}

	return dataLen
}

func (b *binaryLogger) Log(data []byte) {
	if b.currenLogFileSize >= b.logFileMaxSize {
		b.Flush()

		b.bufLock.Lock()
		b.rotateCurrenLogFile()
		b.bufLock.Unlock()
	}

	b.bufLock.Lock()

	var dataLen int64 = 0

	dataLen += b.insertData(data)
	dataLen += b.insertData(RowDelimiter)
	b.currenLogFileSize += dataLen
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
	// todo - err for check nil log file

	b.bufLock.Lock()

	if b.insertsCount > 0 {
		if _, err := b.currentLogFile.Write(b.buf.Bytes()); err != nil {
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

func (b *binaryLogger) Read(offset int64, count int64, whence int) ([]byte, error) {
	buffer := make([]byte, count)

	if err := b.ReadTo(&buffer, offset, whence); err != nil {
		return []byte{}, err
	}

	return buffer, nil
}

func (b *binaryLogger) ReadTo(buffer *[]byte, offset int64, whence int) error {
	_, err := b.currentLogFile.Seek(offset, whence)

	if err != nil {
		return err
	}

	if _, err = b.currentLogFile.Read(*buffer); err != nil {
		return err
	}

	return nil
}

func (b *binaryLogger) DecodeLen(data []byte) int {
	dataLen := 0

	for _, v := range data {
		if v != Space && v != LineBreak && v != EmptyByte {
			dataLen++
		}
	}

	return dataLen
}

func (b *binaryLogger) Decode(data []byte) []byte {
	pure := make([]byte, 0, len(data))

	for _, v := range data {
		if v != Space && v != LineBreak && v != EmptyByte {
			pure = append(pure, v)
		}
	}

	dist := make([]byte, hex.DecodedLen(len(pure)))

	if _, err := hex.Decode(dist, pure); err != nil {
		_, _ = fmt.Fprint(b.errWriter, err.Error(), "\n")
	}

	return dist
}
