package binarylog

import (
	"bufio"
	"errors"
	"fmt"
	"os"
)

func (b *binaryLogger) openLogFile(number int, appendRegistry bool) (*os.File, error) {
	fileName := fmt.Sprintf(logFileTemplate, number)

	if appendRegistry {
		if _, err := b.logFilesRegistry.WriteString(fileName + "\n"); err != nil {
			return nil, errors.New("failed to append in registry file: " + err.Error())
		} else {
			b.logFilesMap[number] = fileName
		}
	}

	filePath := b.basePath + string(os.PathSeparator) + fileName

	return os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
}

func (b *binaryLogger) OpenForRead(number int) (*os.File, error) {
	if _, exists := b.logFilesMap[number]; !exists {
		return nil, ErrLogFileNotExists
	}

	fileName := fmt.Sprintf(logFileTemplate, number)
	filePath := b.basePath + string(os.PathSeparator) + fileName

	return os.OpenFile(filePath, os.O_RDONLY, 0644)
}

func (b *binaryLogger) rotateLogFile() error {
	b.logFilesCount++

	if err := b.logFile.Close(); err != nil {
		return errors.New("failed close old log file: " + err.Error())
	}

	b.logFile = nil
	logFile, err := b.openLogFile(b.logFilesCount, true)

	if err != nil {
		return errors.New("failed to init log file file: " + err.Error())
	}

	b.logFile = logFile
	b.logFileSize = 0
	b.lastLineBytesCount = 0

	return nil
}

func (b *binaryLogger) initLogFile() error {
	if b.logFilesRegistry == nil {
		return errors.New("cant init log file without registry")
	}

	needAppendRegistry := false
	fileName := b.getLastLogFileName()

	if len(fileName) == 0 {
		b.logFilesCount++
		needAppendRegistry = true
	}

	logFile, err := b.openLogFile(b.logFilesCount, needAppendRegistry)

	if err == nil {
		b.logFile = logFile
	} else {
		return errors.New("Failed to init log file file: " + err.Error())
	}

	return nil
}

func (b *binaryLogger) initRegistryFile() error {
	filePath := b.basePath + string(os.PathSeparator) + registryFileName
	registry, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND, 0644)

	if err == nil {
		b.logFilesRegistry = registry
	} else {
		return errors.New("Failed to init registry file: " + err.Error())
	}

	scanner := bufio.NewScanner(b.logFilesRegistry)

	for scanner.Scan() {
		b.logFilesCount++
		b.logFilesMap[b.logFilesCount] = scanner.Text()
	}

	return nil
}

func (b *binaryLogger) calculateLogFileSize() int64 {
	info, _ := b.logFile.Stat()
	return info.Size()
}

func (b *binaryLogger) getLastLogFileName() string {
	return b.logFilesMap[b.logFilesCount]
}

func (b *binaryLogger) SetLogFileMaxSize(size int64) {
	b.logFileMaxSize = size
}

func (b *binaryLogger) CloseLogFile() error {
	return b.logFile.Close()
}

func (b *binaryLogger) Shutdown() {
	_ = b.logFile.Close()
	_ = b.logFilesRegistry.Close()
}

func (b *binaryLogger) logErrorString(err string) {
	_, _ = fmt.Fprint(b.errWriter, err, "\n")
}

func (b *binaryLogger) logError(err error) {
	_, _ = fmt.Fprint(b.errWriter, err.Error(), "\n")
}
