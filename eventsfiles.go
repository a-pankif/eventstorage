package eventstorage

import (
	"bufio"
	"errors"
	"fmt"
	"os"
)

func (s *EventStorage) openEventsFile(number int, appendRegistry bool) (*os.File, error) {
	fileName := s.getFileName(number)
	filePath := s.getFilePath(fileName)

	if appendRegistry {
		if _, err := s.filesRegistry.WriteString(fileName + "\n"); err != nil {
			return nil, errors.New("failed to append in registry file: " + err.Error())
		}
	}

	writeFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		return nil, err
	}

	if _, exists := s.read.readableFiles[number]; !exists {
		readFile, err := os.OpenFile(filePath, os.O_RDONLY, 0644)

		if err != nil {
			_ = writeFile.Close()
			return nil, err
		}

		s.read.readableFiles[number] = readFile
	}

	return writeFile, nil
}

func (s *EventStorage) rotateEventsFile() error {
	if err := s.write.file.Close(); err != nil {
		return errors.New("failed close old events file: " + err.Error())
	}

	s.write.file = nil
	file, err := s.openEventsFile(s.filesCount()+1, true)

	if err != nil {
		return errors.New("rotate failed, open events file err: " + err.Error())
	}

	s.write.file = file
	s.write.fileSize = 0

	return nil
}

func (s *EventStorage) initEventsFile() error {
	if s.filesRegistry == nil {
		return errors.New("cant init events file without registry")
	}

	needAppendRegistry := false
	number := s.filesCount()

	if number == 0 {
		number++
		needAppendRegistry = true
	}

	file, err := s.openEventsFile(number, needAppendRegistry)

	if err == nil {
		s.write.file = file
	} else {
		return errors.New("Failed to init events file: " + err.Error())
	}

	return nil
}

func (s *EventStorage) initFilesRegistry() error {
	filePath := s.getFilePath(registryFileName)
	registry, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)

	if err == nil {
		s.filesRegistry = registry
	} else {
		return errors.New("Failed to init files registry: " + err.Error())
	}

	scanner := bufio.NewScanner(s.filesRegistry)

	for scanner.Scan() {
		path := s.getFilePath(scanner.Text())
		file, err := os.OpenFile(path, os.O_RDONLY, 0644)

		if err != nil {
			return errors.New("Failed to open events file to read: " + err.Error())
		}

		s.read.readableFiles[s.filesCount()+1] = file
	}

	return nil
}

func (s *EventStorage) filesCount() int {
	return len(s.read.readableFiles)
}

func (s *EventStorage) calculateWriteFileSize() int64 {
	info, _ := s.write.file.Stat()
	return info.Size()
}

func (s *EventStorage) SetWriteFileMaxSize(size int64) {
	s.write.fileMaxSize = size
}

func (s *EventStorage) getFileName(number int) string {
	return fmt.Sprintf(eventsFileNameTemplate, number)
}

func (s *EventStorage) getFilePath(fileName string) string {
	return s.basePath + string(os.PathSeparator) + fileName
}

func (s *EventStorage) Shutdown() {
	s.write.locker.Lock()
	s.read.locker.Lock()

	defer func() {
		s.write.locker.Unlock()
		s.read.locker.Unlock()
	}()

	go func() {
		s.turnedOff <- true
	}()

	_ = s.write.file.Close()
	_ = s.filesRegistry.Close()

	for number := 1; number <= s.filesCount(); number++ {
		file, _ := s.read.readableFiles[number]
		_ = file.Close()
	}
}
