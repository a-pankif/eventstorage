package binarylog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"
)

func New(basePath string, errWriter io.Writer) (*eventStorage, error) {
	b := &eventStorage{
		basePath:       basePath,
		buf:            new(bytes.Buffer),
		encodeBuf:      make([]byte, 3),
		logFilesMap:    make(logFilesMap),
		logFilesCount:  0,
		logFileMaxSize: 100 * MB,
		errWriter:      errWriter,
	}

	if err := b.initRegistryFile(); err != nil {
		return nil, err
	}

	if err := b.initLogFile(); err != nil {
		return nil, err
	}

	b.logFileSize = b.calculateLogFileSize()

	return b, nil
}

func (b *eventStorage) Log(data []byte) (writtenLen int64, err error) {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buf.Write(data)
	b.buf.WriteByte(LineBreak)

	writtenLen += int64(len(data) + 1)

	b.logFileSize += writtenLen
	b.insertsCount++

	if b.autoFlushCount > 0 && b.insertsCount >= b.autoFlushCount {
		if _, err = b.flush(); err != nil {
			return
		}
	}

	if b.logFileSize >= b.logFileMaxSize {
		if _, err = b.flush(); err != nil {
			return
		}

		if err = b.rotateLogFile(); err != nil {
			return
		}
	}

	return
}

func (b *eventStorage) flush() (count int, err error) {
	if b.insertsCount > 0 {
		if _, err := b.logFile.Write(b.buf.Bytes()); err != nil {
			return 0, errors.New("flush failed: " + err.Error())
		} else {
			b.buf.Truncate(0)
			count = b.insertsCount
			b.insertsCount = 0
		}
	}

	return
}

func (b *eventStorage) Flush() (count int, err error) {
	b.locker.Lock()
	defer b.locker.Unlock()
	return b.flush()
}

func (b *eventStorage) SetAutoFlushCount(count int) {
	b.autoFlushCount = count
}

func (b *eventStorage) GetAutoFlushCount() int {
	return b.autoFlushCount
}

func (b *eventStorage) SetAutoFlushTime(period time.Duration) error {
	if b.autoFlushTime != 0 {
		return ErrAutoFlushTimeAlreadySet
	}

	if period <= 0 {
		return ErrAutoFlushTimeTooLow
	}

	b.autoFlushTime = period

	go func() {
		for range time.Tick(period) {
			// todo - support Shutdown function to exit from gorutine
			if _, err := b.Flush(); err != nil {
				b.logErrorString("time flush failed: " + err.Error())
			}
		}
	}()

	return nil
}

func (b *eventStorage) Read(offset int64, count int64, whence int) ([]byte, error) {
	buffer := make([]byte, count)

	if err := b.ReadTo(&buffer, offset, whence); err != nil {
		return []byte{}, err
	}

	return buffer, nil
}

func (b *eventStorage) ReadTo(buffer *[]byte, offset int64, whence int) error {
	_, err := b.logFile.Seek(offset, whence)

	if err != nil {
		return err
	}

	if _, err = b.logFile.Read(*buffer); err != nil {
		return err
	}

	return nil
}

func (b *eventStorage) ReadEvents(count int64, offset int64) {
	var seekOffset int64 = 0
	events := make([][]byte, 0, count)
	readBuffer := make([]byte, lineLength)

	for number := 1; number <= b.logFilesCount; number++ {
		file, _ := b.OpenForRead(number)
		emptyBytesCount := 0
		event := new(bytes.Buffer)

		for {
			_, _ = file.Seek(seekOffset, 0)
			readCount, err := file.Read(readBuffer)

			if err == io.EOF {
				seekOffset = 0
				_ = file.Close()
				break
			}

			fmt.Println(readBuffer, LineBreak, EmptyByte, string(EmptyByte))
			for i := 0; i < readCount; i++ {
				v := readBuffer[i]
				if v == 48 {
					emptyBytesCount++
				} else if v != Space && v != LineBreak {
					emptyBytesCount = 0
				}

				fmt.Println(v, emptyBytesCount)
				if v != Space && v != LineBreak && v != 48 {
					event.WriteByte(v)
				}
			}

			// fmt.Println(readCount, string(readBuffer[0:readCount]))
			events = append(events, readBuffer[0:readCount])
			seekOffset += lineLength
		}

		fmt.Println(string(event.Bytes()))
	}

	// fmt.Println(events)

}
