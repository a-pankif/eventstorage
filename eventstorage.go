package eventstorage

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"time"
)

func New(basePath string) (*eventStorage, error) {
	s := &eventStorage{
		basePath:  basePath,
		write:     &write{buf: new(bytes.Buffer), fileMaxSize: 100 * MB},
		read:      &read{readableFiles: make(readableFiles), buf: new(strings.Builder), readBuf: make([]byte, readBufLimit)},
		turnedOff: make(chan bool, 1),
	}

	if err := s.initFilesRegistry(); err != nil {
		return nil, err
	}

	if err := s.initEventsFile(); err != nil {
		return nil, err
	}

	s.write.fileSize = s.calculateWriteFileSize()

	return s, nil
}

func (s *eventStorage) Write(data []byte) (writtenLen int64, err error) {
	s.write.locker.Lock()
	defer s.write.locker.Unlock()

	s.write.buf.Write(data)
	s.write.buf.WriteByte(LineBreak)

	writtenLen += int64(len(data) + 1)

	s.write.fileSize += writtenLen
	s.write.insertsCount++

	if s.write.autoFlushCount > 0 && s.write.insertsCount >= s.write.autoFlushCount {
		if _, err = s.flush(); err != nil {
			return
		}
	}

	if s.write.fileSize >= s.write.fileMaxSize {
		if _, err = s.flush(); err != nil {
			return
		}

		if err = s.rotateEventsFile(); err != nil {
			return
		}
	}

	return
}

func (s *eventStorage) flush() (count int, err error) {
	if s.write.insertsCount > 0 {
		if _, err = s.write.file.Write(s.write.buf.Bytes()); err != nil {
			return 0, errors.New("flush failed: " + err.Error())
		} else {
			s.write.buf.Truncate(0)
			count = s.write.insertsCount
			s.write.insertsCount = 0
		}
	}

	return
}

func (s *eventStorage) Flush() (count int, err error) {
	s.write.locker.Lock()
	defer s.write.locker.Unlock()
	return s.flush()
}

func (s *eventStorage) SetAutoFlushCount(count int) {
	s.write.autoFlushCount = count
}

func (s *eventStorage) GetAutoFlushCount() int {
	return s.write.autoFlushCount
}

func (s *eventStorage) SetAutoFlushTime(period time.Duration) error {
	if period <= 0 {
		return ErrAutoFlushTimeTooLow
	}

	if s.write.autoFlushTime != 0 {
		return ErrAutoFlushTimeAlreadySet
	}

	s.write.autoFlushTime = period

	go func() {
		for range time.Tick(period) {
			select {
			case <-s.turnedOff:
				return
			default:
			}

			_, _ = s.Flush()
		}
	}()

	return nil
}

func (s *eventStorage) ReadTo(count int, offset int, events *[]string) {
	s.read.locker.Lock()
	defer s.read.locker.Unlock()

	s.read.eventsCount = 0
	s.read.eventsSaved = 0
	s.read.buf.Reset()

	for number := 1; number <= s.filesCount(); number++ {
		file := s.read.readableFiles[number]
		s.read.seekOffset = 0

		for {
			_, _ = file.Seek(s.read.seekOffset, 0)
			readCount, err := file.Read(s.read.readBuf)

			if err != nil && (err == io.EOF || strings.Contains(err.Error(), "file already closed")) {
				break
			}

			for i := 0; i < readCount; i++ {
				if s.read.readBuf[i] == LineBreak {
					if offset <= s.read.eventsCount {
						*events = append(*events, s.read.buf.String())
						s.read.eventsSaved++
					}

					s.read.eventsCount++
					s.read.buf.Reset()

					if s.read.eventsSaved == count {
						return
					}
				} else {
					s.read.buf.WriteByte(s.read.readBuf[i])
				}
			}

			s.read.seekOffset += readBufLimit
		}
	}
}

func (s *eventStorage) Read(count int, offset int) []string {
	events := make([]string, 0, count)
	s.ReadTo(count, offset, &events)
	return events
}
