package binarylog

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func (b *binaryLogger) calculateCurrenLogFileSize() int64 {
	info, _ := b.currentLogFile.Stat()
	return info.Size()
}

func (b *binaryLogger) openLogFile(number int, appendRegistry bool) (*os.File, error) {
	currentLogFileName := fmt.Sprintf(logFileTemplate, number)

	if appendRegistry {
		b.registryAppend(currentLogFileName)
	}

	filePath := b.basePath + string(os.PathSeparator) + currentLogFileName

	return os.OpenFile(filePath, os.O_CREATE|os.O_APPEND, 0644)
}

func (b *binaryLogger) rotateCurrenLogFile() {
	b.logFilesCount++

	logFile, err := b.openLogFile(b.logFilesCount, true)

	if err == nil {
		b.currentLogFile = logFile
		b.currenLogFileSize = 0
		b.lastLineBytesCount = 0
	} else {
		b.currentLogFile = nil
		b.logErrorString("Failed to init log file file: " + err.Error())
	}
}

func (b *binaryLogger) initCurrenLogFile() {
	if b.logFilesRegistry == nil {
		b.logErrorString("Cant init log file without registry.")
		return
	}

	reader := bufio.NewReader(b.logFilesRegistry)
	currentLogFileName := ""

	for {
		line, err := reader.ReadString('\n')

		if err != nil && err != io.EOF {
			b.logErrorString("Error while parse log registry: " + err.Error())
			break
		}

		if len(line) > 2 { // 2 - is length of single line break
			currentLogFileName = line
			b.logFilesCount++
		}

		if err == io.EOF {
			break
		}
	}

	needAppendRegistry := false

	if len(currentLogFileName) == 0 {
		b.logFilesCount++
		needAppendRegistry = true
	}

	logFile, err := b.openLogFile(b.logFilesCount, needAppendRegistry)

	if err == nil {
		b.currentLogFile = logFile
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
	}
}

func (b *binaryLogger) registryAppend(logName string) {
	if _, err := b.logFilesRegistry.WriteString(logName + "\n"); err != nil {
		b.logErrorString("Failed to append in registry file: " + err.Error())
	}
}

func (b *binaryLogger) SetLogFileSize(size int64) {
	b.logFileMaxSize = size
}

func (b *binaryLogger) CloseLogFile() error {
	return b.currentLogFile.Close()
}

func (b *binaryLogger) logErrorString(err string) {
	_, _ = fmt.Fprint(b.errWriter, err, "\n")
}

func (b *binaryLogger) logError(err error) {
	_, _ = fmt.Fprint(b.errWriter, err.Error(), "\n")
}
