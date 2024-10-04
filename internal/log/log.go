package log

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

const timeFormat = "2006-01-02T15:04:05Z07:00"

func parseLevel(level string) (Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return DEBUG, nil
	case "info":
		return INFO, nil
	case "warn":
		return WARN, nil
	case "warning":
		return WARN, nil
	case "err":
		return ERROR, nil
	case "error":
		return ERROR, nil
	case "fatal":
		return FATAL, nil
	default:
		return 0, fmt.Errorf("unknown log level %q", level)
	}
}

func (lvl Level) String() string {
	switch lvl {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return ""
	}
}

type Logger struct {
	path string
	fh   *os.File
	m    sync.Mutex
	lvl  Level
}

func Open(path, level string) (*Logger, error) {
	l := &Logger{path: path}
	lvl, err := parseLevel(level)
	if err != nil {
		return nil, err
	}
	l.lvl = lvl
	err = l.ReOpen()
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (l *Logger) ReOpen() error {
	l.m.Lock()
	defer l.m.Unlock()
	if l.path == "" || l.path == "/dev/stderr" {
		l.fh = os.Stderr
		return nil
	}
	if l.fh != nil {
		_ = l.fh.Close()
	}
	fh, err := os.OpenFile(l.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log %s: %w", l.path, err)
	}
	l.fh = fh
	return nil
}

func (l *Logger) ReOpenOnSignal(sig syscall.Signal) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, sig)
	go func() {
		for {
			<-sigs
			err := l.ReOpen()
			if err != nil {
				log.Printf("failed to reopen log file: %v", err)
			}
		}
	}()
}

func (l *Logger) printf(lvl Level, msg string, args ...interface{}) {
	if lvl < l.lvl {
		return
	}
	data := fmt.Sprintf("%s %s: ", time.Now().Format(timeFormat), lvl) + fmt.Sprintf(msg, args...) + "\n"
	l.m.Lock()
	_, _ = l.fh.Write([]byte(data))
	l.m.Unlock()
}

func (l *Logger) Debug(msg string) {
	l.Debugf("%s", msg)
}

func (l *Logger) Info(msg string) {
	l.Infof("%s", msg)
}

func (l *Logger) Warn(msg string) {
	l.Warnf("%s", msg)
}

func (l *Logger) Error(msg string) {
	l.Errorf("%s", msg)
}

func (l *Logger) Fatal(msg string) {
	l.Fatalf("%s", msg)
}

func (l *Logger) Debugf(msg string, args ...interface{}) {
	l.printf(DEBUG, msg, args...)
}

func (l *Logger) Infof(msg string, args ...interface{}) {
	l.printf(INFO, msg, args...)
}

func (l *Logger) Warnf(msg string, args ...interface{}) {
	l.printf(WARN, msg, args...)
}

func (l *Logger) Errorf(msg string, args ...interface{}) {
	l.printf(ERROR, msg, args...)
}

func (l *Logger) Fatalf(msg string, args ...interface{}) {
	l.printf(FATAL, msg, args...)
}

type Prefix string

const (
	RepairPrefix   = "repair"
	RecoveryPrefix = "recovery"
)

type PrefixLogger struct {
	prefix       string
	sourceLogger *Logger
}

func NewPrefixLogger(logger *Logger, prefix string) *PrefixLogger {
	return &PrefixLogger{
		prefix:       prefix,
		sourceLogger: logger,
	}
}

func (l *PrefixLogger) Debug(msg string) {
	l.sourceLogger.Debugf("%s: %s", l.prefix, msg)
}

func (l *PrefixLogger) Info(msg string) {
	l.sourceLogger.Infof("%s: %s", l.prefix, msg)
}

func (l *PrefixLogger) Warn(msg string) {
	l.sourceLogger.Warnf("%s: %s", l.prefix, msg)
}

func (l *PrefixLogger) Error(msg string) {
	l.sourceLogger.Errorf("%s: %s", l.prefix, msg)
}

func (l *PrefixLogger) Fatal(msg string) {
	l.sourceLogger.Fatalf("%s: %s", l.prefix, msg)
}

func (l *PrefixLogger) Debugf(msg string, args ...interface{}) {
	msg = fmt.Sprintf("%s: %s", l.prefix, msg)
	l.sourceLogger.Debugf(msg, args...)
}

func (l *PrefixLogger) Infof(msg string, args ...interface{}) {
	msg = fmt.Sprintf("%s: %s", l.prefix, msg)
	l.sourceLogger.Infof(msg, args...)
}

func (l *PrefixLogger) Warnf(msg string, args ...interface{}) {
	msg = fmt.Sprintf("%s: %s", l.prefix, msg)
	l.sourceLogger.Warnf(msg, args...)
}

func (l *PrefixLogger) Errorf(msg string, args ...interface{}) {
	msg = fmt.Sprintf("%s: %s", l.prefix, msg)
	l.sourceLogger.Errorf(msg, args...)
}

func (l *PrefixLogger) Fatalf(msg string, args ...interface{}) {
	msg = fmt.Sprintf("%s: %s", l.prefix, msg)
	l.sourceLogger.Fatalf(msg, args...)
}
