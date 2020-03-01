package main

import (
	"io"
	"log"
	"os"
)

const (
	// DEBUG log level
	DEBUG = iota
	// INFO log level
	INFO
	// WARNING log level
	WARNING
	// ERROR log level
	ERROR
	// CRITICAL log level
	CRITICAL
	// FATAL log level
	FATAL
)

var levelToPrefix = map[int]string{
	DEBUG:    "DEBUG ",
	INFO:     "INFO ",
	WARNING:  "WARNING ",
	ERROR:    "ERROR ",
	CRITICAL: "CRITICAL ",
	FATAL:    "FATAL ",
}

// Logger is the custom logger.
type Logger struct {
	*log.Logger
	level int
}

// NewLogger returns logger which writes to specified files.
func NewLogger(config *LogConfig) *Logger {
	prefix := "INFO "
	var level int
	if config.Level == "" {
		level = INFO
	} else {
		switch config.Level {
		case "DEBUG":
			level = DEBUG
		case "INFO":
			level = INFO
		case "WARNING":
			level = WARNING
		case "ERROR":
			level = ERROR
		case "CRITICAL":
			level = CRITICAL
		case "FATAL":
			level = FATAL
		default:
			log.Fatalf("Invalid log level in config: %s", config.Level)
		}
	}
	flag := log.LstdFlags

	if len(config.Files) == 0 {
		return &Logger{Logger: log.New(os.Stderr, prefix, flag), level: level}
	}

	outs := make([]io.Writer, 0)
	for _, file := range config.Files {
		if file == "stderr" {
			outs = append(outs, os.Stderr)
		} else {
			path := getPath(file)
			out, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			checkErr(err)
			outs = append(outs, out)
		}
	}

	if len(outs) == 1 {
		return &Logger{Logger: log.New(outs[0], prefix, flag), level: level}
	}

	return &Logger{Logger: log.New(io.MultiWriter(outs...), prefix, flag), level: level}
}

// printLevel logs message with a given level.
func (l *Logger) printLevel(msg string, level int, args ...interface{}) {
	if level < l.level {
		return
	}
	l.SetPrefix(levelToPrefix[level])
	if len(args) == 0 {
		l.Logger.Print(msg)
	} else {
		l.Logger.Printf(msg, args...)
	}
}

// Debug logs message with debug level.
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.printLevel(msg, DEBUG, args...)
}

// Info logs message with info level.
func (l *Logger) Info(msg string, args ...interface{}) {
	l.printLevel(msg, INFO, args...)
}

// Warning logs message with warning level.
func (l *Logger) Warning(msg string, args ...interface{}) {
	l.printLevel(msg, WARNING, args...)
}

// Error logs message with error level.
func (l *Logger) Error(msg string, args ...interface{}) {
	l.printLevel(msg, ERROR, args...)
}

// Critical logs message with critical level.
func (l *Logger) Critical(msg string, args ...interface{}) {
	l.printLevel(msg, CRITICAL, args...)
}

// Fatal sets prefix and calls log.Fatal.
func (l *Logger) Fatal(v ...interface{}) {
	l.SetPrefix(levelToPrefix[FATAL])
	l.Logger.Fatal(v...)
}

// Fatalln sets prefix and calls log.Fatalln.
func (l *Logger) Fatalln(v ...interface{}) {
	l.SetPrefix(levelToPrefix[FATAL])
	l.Logger.Fatalln(v...)
}

// Fatalf sets prefix and calls log.Fatalf.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.SetPrefix(levelToPrefix[FATAL])
	l.Logger.Fatalf(format, v...)
}

// Print merely prints without prefix and flags.
func (l *Logger) Print(v ...interface{}) {
	l.SetPrefix("")
	flags := l.Flags()
	l.SetFlags(0)
	l.Logger.Print(v...)
	l.SetFlags(flags)
}

// Println merely prints without prefix and flags.
func (l *Logger) Println(v ...interface{}) {
	l.SetPrefix("")
	flags := l.Flags()
	l.SetFlags(0)
	l.Logger.Println(v...)
	l.SetFlags(flags)
}

// Printf merely prints without prefix and flags.
func (l *Logger) Printf(format string, v ...interface{}) {
	l.SetPrefix("")
	flags := l.Flags()
	l.SetFlags(0)
	l.Logger.Printf(format, v...)
	l.SetFlags(flags)
}
