package binarylog

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"time"
)

func New(basePath string, errWriter io.Writer) (*eventStorage, error) {
	b := &eventStorage{
		basePath:           basePath,
		buf:                new(bytes.Buffer),
		eventsFilesMap:     make(logFilesMap),
		eventsFilesReadMap: make(eventsFileRead),
		filesCount:         0,
		fileMaxSize:        100 * MB,
		errWriter:          errWriter,
	}

	if err := b.initRegistryFile(); err != nil {
		return nil, err
	}

	if err := b.initEventsFile(); err != nil {
		return nil, err
	}

	b.eventsFileSize = b.calculateLogFileSize()

	return b, nil
}

func (b *eventStorage) Log(data []byte) (writtenLen int64, err error) {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buf.Write(data)
	b.buf.WriteByte(LineBreak)

	writtenLen += int64(len(data) + 1)

	b.eventsFileSize += writtenLen
	b.insertsCount++

	if b.autoFlushCount > 0 && b.insertsCount >= b.autoFlushCount {
		if _, err = b.flush(); err != nil {
			return
		}
	}

	if b.eventsFileSize >= b.fileMaxSize {
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
	if b.insertsCount > 0 {
		if _, err := b.eventsFile.Write(b.buf.Bytes()); err != nil {
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
	if period <= 0 {
		return ErrAutoFlushTimeTooLow
	}

	if b.autoFlushTime != 0 {
		return ErrAutoFlushTimeAlreadySet // todo - supports cancel curren gorutine by channel and set up new
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
	_, err := b.eventsFile.Seek(offset, whence)

	if err != nil {
		return err
	}

	if _, err = b.eventsFile.Read(*buffer); err != nil {
		return err
	}

	return nil
}

func (b *eventStorage) ReadEvents(count int, offset int) []string {
	var seekOffset int64 = 0
	buf := new(strings.Builder)
	events := make([]string, 0, count)
	readBuffer := make([]byte, 100)
	eventsSaved := 0
	eventsCount := 0

	for number := 1; number <= b.filesCount; number++ {
		file, _ := b.OpenForRead(number)

		for {
			_, _ = file.Seek(seekOffset, 0)
			readCount, err := file.Read(readBuffer)

			if err == io.EOF {
				seekOffset = 0
				break
			}

			for i := 0; i < readCount; i++ {
				if readBuffer[i] == LineBreak {
					if offset <= eventsCount {
						events = append(events, buf.String())
						eventsSaved++
					}

					eventsCount++
					buf.Reset()

					if eventsSaved == count {
						return events
					}
				} else {
					buf.WriteByte(readBuffer[i])
				}
			}

			seekOffset += 100
		}
	}

	return events
}
