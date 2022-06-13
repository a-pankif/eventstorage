package eventstorage

import (
	"bufio"
	"errors"
	"fmt"
	"os"
)

func (b *eventStorage) openEventsFile(number int, appendRegistry bool) (*os.File, error) {
	fileName := b.getFileName(number)
	filePath := b.getFilePath(number)

	if appendRegistry {
		if _, err := b.filesRegistry.WriteString(fileName + "\n"); err != nil {
			return nil, errors.New("failed to append in registry file: " + err.Error())
		}
	}

	writeFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		return nil, err
	}

	if _, exists := b.read.readableFiles[number]; !exists {
		readFile, err := os.OpenFile(filePath, os.O_RDONLY, 0644)

		if err != nil {
			_ = writeFile.Close()
			return nil, err
		}

		b.read.readableFiles[number] = readFile
	}

	return writeFile, nil
}

func (b *eventStorage) rotateEventsFile() error {
	if err := b.write.file.Close(); err != nil {
		return errors.New("failed close old log file: " + err.Error())
	}

	b.write.file = nil
	logFile, err := b.openEventsFile(b.filesCount()+1, true)

	if err != nil {
		return errors.New("failed to init log file file: " + err.Error())
	}

	b.write.file = logFile
	b.write.fileSize = 0

	return nil
}

func (b *eventStorage) initEventsFile() error {
	if b.filesRegistry == nil {
		return errors.New("cant init log file without registry")
	}

	needAppendRegistry := false

	fileName := b.getLastLogFileName()
	number := b.filesCount()

	if len(fileName) == 0 {
		number++
		needAppendRegistry = true
	}

	logFile, err := b.openEventsFile(number, needAppendRegistry)

	if err == nil {
		b.write.file = logFile
	} else {
		return errors.New("Failed to init log file file: " + err.Error())
	}

	return nil
}

func (b *eventStorage) initRegistryFile() error {
	filePath := b.basePath + string(os.PathSeparator) + registryFileName
	registry, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND, 0644)

	if err == nil {
		b.filesRegistry = registry
	} else {
		return errors.New("Failed to init registry file: " + err.Error())
	}

	scanner := bufio.NewScanner(b.filesRegistry)

	for scanner.Scan() {
		path := b.basePath + string(os.PathSeparator) + scanner.Text()
		file, err := os.OpenFile(path, os.O_RDONLY, 0644)

		if err != nil {
			return err
		}

		b.read.readableFiles[b.filesCount()+1] = file
	}

	return nil
}

func (b *eventStorage) filesCount() int {
	return len(b.read.readableFiles)
}

func (b *eventStorage) calculateLogFileSize() int64 {
	info, _ := b.write.file.Stat()
	return info.Size()
}

func (b *eventStorage) getLastLogFileName() string {
	if b.filesCount() == 0 {
		return ""
	}

	return b.read.readableFiles[b.filesCount()].Name()
}

func (b *eventStorage) SetLogFileMaxSize(size int64) {
	b.write.fileMaxSize = size
}

func (b *eventStorage) getFileName(number int) string {
	return fmt.Sprintf(eventsFileNameTemplate, number)
}

func (b *eventStorage) getFilePath(number int) string {
	return b.basePath + string(os.PathSeparator) + b.getFileName(number)
}

func (b *eventStorage) Shutdown() {
	b.write.locker.Lock()
	b.read.locker.Lock()

	defer func() {
		b.write.locker.Unlock()
		b.read.locker.Unlock()
	}()

	go func() {
		b.turnedOff <- true
	}()

	_ = b.write.file.Close()
	_ = b.filesRegistry.Close()

	for number := 1; number <= b.filesCount(); number++ {
		file, _ := b.read.readableFiles[number]
		_ = file.Close()
	}
}
