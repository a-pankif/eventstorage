package binarylog

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
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

	return os.OpenFile(filePath, os.O_CREATE|os.O_APPEND, 0644)
}

func (b *binaryLogger) rotateLogFile() {
	b.logFilesCount++

	if err := b.logFile.Close(); err != nil {
		b.logErrorString("Failed close old log file: " + err.Error())
	}

	logFile, err := b.openLogFile(b.logFilesCount, true)

	if err == nil {
		b.logFile = logFile
		b.logFileSize = 0
		b.lastLineBytesCount = 0
	} else {
		b.logFile = nil
		b.logErrorString("Failed to init log file file: " + err.Error())
	}
}

func (b *binaryLogger) initLogFile() {
	if b.logFilesRegistry == nil {
		b.logErrorString("Cant init log file without registry.")
		return
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
		b.logErrorString("Failed to init log file file: " + err.Error())
	}
}

func (b *binaryLogger) initRegistryFile() {
	filePath := b.basePath + string(os.PathSeparator) + registryFileName
	registry, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND, 0644)

	if err == nil {
		b.logFilesRegistry = registry
	} else {
		b.logErrorString("Failed to init registry file: " + err.Error())
		return
	}

	reader := bufio.NewReader(b.logFilesRegistry)

	for {
		line, err := reader.ReadString(LineBreak)
		if err != nil && err != io.EOF {
			b.logErrorString("Error while parse log registry: " + err.Error())
			break
		}

		line = strings.NewReplacer("\n", "", "\r", "").Replace(line)

		if len(line) > 0 {
			b.logFilesCount++
			b.logFilesMap[b.logFilesCount] = line
		}

		if err == io.EOF {
			break
		}
	}
}

func (b *binaryLogger) calculateLogFileSize() int64 {
	info, _ := b.logFile.Stat()
	return info.Size()
}

func (b *binaryLogger) getLastLogFileName() string {
	return b.logFilesMap[b.logFilesCount]
}

func (b *binaryLogger) SetLogFileSize(size int64) {
	b.logFileMaxSize = size
}

func (b *binaryLogger) CloseLogFile() error {
	return b.logFile.Close()
}

func (b *binaryLogger) logErrorString(err string) {
	_, _ = fmt.Fprint(b.errWriter, err, "\n")
}

func (b *binaryLogger) logError(err error) {
	_, _ = fmt.Fprint(b.errWriter, err.Error(), "\n")
}
