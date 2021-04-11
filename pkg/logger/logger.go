package logger

import (
	"log"
	"os"
)

func init() {
	New(10)
}

const (
	OFF   = 0
	FATAL = 1
	ERROR = 2
	INFO  = 4
	DEBUG = 5
)

type Writer struct {
	infoLogger  *log.Logger
	debugLogger *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
	level       int
}

var logger *Writer

func New(level int) {
	logger = &Writer{
		infoLogger:  log.New(os.Stdout, "[Info]", log.LstdFlags),
		debugLogger: log.New(os.Stdout, "[Debug]", log.LstdFlags),
		errorLogger: log.New(os.Stderr, "[Error]", log.LstdFlags),
		fatalLogger: log.New(os.Stderr, "[Fatal]", log.LstdFlags),
		level:       level,
	}
}

func Info(std ...interface{}) {
	if logger.level >= INFO {
		logger.infoLogger.Println(std...)
	}
}

func Debug(std ...interface{}) {
	if logger.level >= DEBUG {
		logger.debugLogger.Println(std...)
	}
}

func Error(std ...interface{}) {
	if logger.level >= ERROR {
		logger.errorLogger.Println(std...)
	}
}

func Fatal(std ...interface{}) {
	if logger.level >= FATAL {
		logger.fatalLogger.Fatalln(std...)
	}
}
