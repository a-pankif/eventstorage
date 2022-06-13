package eventstorage

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"time"
)

func New(basePath string) (*eventStorage, error) {
	b := &eventStorage{
		basePath:  basePath,
		write:     &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:      &read{readableFiles: make(readableFiles), buf: new(strings.Builder)},
		turnedOff: make(chan bool, 1),
	}

	if err := b.initRegistryFile(); err != nil {
		return nil, err
	}

	if err := b.initEventsFile(); err != nil {
		return nil, err
	}

	b.write.fileSize = b.calculateLogFileSize()

	return b, nil
}

func (b *eventStorage) Log(data []byte) (writtenLen int64, err error) {
	b.write.locker.Lock()
	defer b.write.locker.Unlock()

	b.write.buf.Write(data)
	b.write.buf.WriteByte(LineBreak)

	writtenLen += int64(len(data) + 1)

	b.write.fileSize += writtenLen
	b.write.insertsCount++

	if b.write.autoFlushCount > 0 && b.write.insertsCount >= b.write.autoFlushCount {
		if _, err = b.flush(); err != nil {
			return
		}
	}

	if b.write.fileSize >= b.write.fileMaxSize {
		if _, err = b.flush(); err != nil {
			return
		}

		if err = b.rotateEventsFile(); err != nil {
			return
		}
	}

	return
}

func (b *eventStorage) flush() (count int, err error) {
	if b.write.insertsCount > 0 {
		if _, err := b.write.file.Write(b.write.buf.Bytes()); err != nil {
			return 0, errors.New("flush failed: " + err.Error())
		} else {
			b.write.buf.Truncate(0)
			count = b.write.insertsCount
			b.write.insertsCount = 0
		}
	}

	return
}

func (b *eventStorage) Flush() (count int, err error) {
	b.write.locker.Lock()
	defer b.write.locker.Unlock()
	return b.flush()
}

func (b *eventStorage) setAutoFlushCount(count int) {
	b.write.autoFlushCount = count
}

func (b *eventStorage) GetAutoFlushCount() int {
	return b.write.autoFlushCount
}

func (b *eventStorage) SetAutoFlushTime(period time.Duration) error {
	if period <= 0 {
		return ErrAutoFlushTimeTooLow
	}

	if b.write.autoFlushTime != 0 {
		return ErrAutoFlushTimeAlreadySet // todo - supports cancel curren gorutine by channel and set up new
	}

	b.write.autoFlushTime = period

	go func() {
		for range time.Tick(period) {
			select {
			case <-b.turnedOff:
				return
			default:
			}

			_, _ = b.Flush()
		}
	}()

	return nil
}

func (b *eventStorage) ReadEvents(count int, offset int) []string {
	b.read.locker.Lock()
	defer b.read.locker.Unlock()

	events := make([]string, 0, count)
	readBuffer := make([]byte, readEventsOpLimit)

	b.read.eventsCount = 0
	b.read.eventsSaved = 0
	b.read.seekOffset = 0
	b.read.buf.Reset()

	for number := 1; number <= b.filesCount(); number++ {
		file := b.read.readableFiles[number]

		for {
			_, _ = file.Seek(b.read.seekOffset, 0)
			readCount, err := file.Read(readBuffer)

			if err != nil && (err == io.EOF || strings.Contains(err.Error(), "file already closed")) {
				b.read.seekOffset = 0
				break
			}

			for i := 0; i < readCount; i++ {
				if readBuffer[i] == LineBreak {
					if offset <= b.read.eventsCount {
						events = append(events, b.read.buf.String())
						b.read.eventsSaved++
					}

					b.read.eventsCount++
					b.read.buf.Reset()

					if b.read.eventsSaved == count {
						return events
					}
				} else {
					b.read.buf.WriteByte(readBuffer[i])
				}
			}

			b.read.seekOffset += readEventsOpLimit
		}
	}

	return events
}
