package binarylog

import (
	"bufio"
	"errors"
	"fmt"
	"os"
)

func (b *eventStorage) openEventsFile(number int, appendRegistry bool) (*os.File, error) {
	fileName := fmt.Sprintf(eventsFileNameTemplate, number)

	if appendRegistry {
		if _, err := b.eventsFilesRegistry.WriteString(fileName + "\n"); err != nil {
			return nil, errors.New("failed to append in registry file: " + err.Error())
		} else {
			b.eventsFilesMap[number] = fileName
		}
	}

	filePath := b.basePath + string(os.PathSeparator) + fileName

	return os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
}

func (b *eventStorage) OpenForRead(number int) (*os.File, error) {
	if _, exists := b.eventsFilesMap[number]; !exists {
		return nil, ErrEventsFileNotExists
	}

	if file, exists := b.eventsFilesReadMap[number]; exists {
		return file, nil
	}

	fileName := fmt.Sprintf(eventsFileNameTemplate, number)
	filePath := b.basePath + string(os.PathSeparator) + fileName

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0644)

	if err != nil {
		return nil, err
	}

	b.eventsFilesReadMap[number] = file

	return file, nil
}

func (b *eventStorage) rotateEventsFile() error {
	b.filesCount++

	if err := b.eventsFile.Close(); err != nil {
		return errors.New("failed close old log file: " + err.Error())
	}

	b.eventsFile = nil
	logFile, err := b.openEventsFile(b.filesCount, true)

	if err != nil {
		return errors.New("failed to init log file file: " + err.Error())
	}

	b.eventsFile = logFile
	b.eventsFileSize = 0

	return nil
}

func (b *eventStorage) initEventsFile() error {
	if b.eventsFilesRegistry == nil {
		return errors.New("cant init log file without registry")
	}

	needAppendRegistry := false
	fileName := b.getLastLogFileName()

	if len(fileName) == 0 {
		b.filesCount++
		needAppendRegistry = true
	}

	logFile, err := b.openEventsFile(b.filesCount, needAppendRegistry)

	if err == nil {
		b.eventsFile = logFile
	} else {
		return errors.New("Failed to init log file file: " + err.Error())
	}

	return nil
}

func (b *eventStorage) initRegistryFile() error {
	filePath := b.basePath + string(os.PathSeparator) + registryFileName
	registry, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND, 0644)

	if err == nil {
		b.eventsFilesRegistry = registry
	} else {
		return errors.New("Failed to init registry file: " + err.Error())
	}

	scanner := bufio.NewScanner(b.eventsFilesRegistry)

	for scanner.Scan() {
		b.filesCount++
		b.eventsFilesMap[b.filesCount] = scanner.Text()
	}

	return nil
}

func (b *eventStorage) calculateLogFileSize() int64 {
	info, _ := b.eventsFile.Stat()
	return info.Size()
}

func (b *eventStorage) getLastLogFileName() string {
	return b.eventsFilesMap[b.filesCount]
}

func (b *eventStorage) SetLogFileMaxSize(size int64) {
	b.fileMaxSize = size
}

func (b *eventStorage) CloseLogFile() error {
	return b.eventsFile.Close()
}

func (b *eventStorage) Shutdown() {
	_ = b.eventsFile.Close()
	_ = b.eventsFilesRegistry.Close()

	for number := 1; number <= b.filesCount; number++ {
		file, _ := b.OpenForRead(number)
		_ = file.Close()
	}
}

func (b *eventStorage) logErrorString(err string) {
	_, _ = fmt.Fprint(b.errWriter, err, "\n")
}

func (b *eventStorage) logError(err error) {
	_, _ = fmt.Fprint(b.errWriter, err.Error(), "\n")
}
