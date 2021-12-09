package main

import (
	"encoding/json"
	"log/syslog"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	levelMap      map[logrus.Level]syslog.Priority
	syslogNameMap map[syslog.Priority]string

	protectedFields map[string]bool

	// DefaultLevel is the default syslog level to use if the logrus level does not map to a syslog level
	DefaultLevel = syslog.LOG_INFO
)

func init() {

	protectedFields = map[string]bool{
		"loggerName":  true,
		"message":     true,
		"timestamp":   true,
		"level":       true,
		"callFuntion": true,
		"error":       true,
		"fileName":    false,
		"line":        false,
		"function":    false,
	}
}

type customFormatter struct {
	loggerName string
}

func NewCustomFormatter(loggerName string) customFormatter {
	return customFormatter{loggerName: loggerName}
}

// Format implements logrus formatter
func (f customFormatter) Format(entry *logrus.Entry) ([]byte, error) {

	gelfEntry := map[string]interface{}{
		"message":    entry.Message,
		"level":      entry.Level,
		"timestamp":  toEpochUnixTime(entry.Time), //toTimestamp
		"loggerName": f.loggerName,
	}

	// if pc, file, line, ok := runtime.Caller(5); ok {
	// 	filename := file[strings.LastIndex(file, "/")+1:]
	// 	gelfEntry["fileName"] = filename
	// 	gelfEntry["line"] = line
	// 	funcname := runtime.FuncForPC(pc).Name()
	// 	gelfEntry["function"] = funcname
	// }

	contextMap := map[string]interface{}{}
	errorMap := map[string]interface{}{}
	for key, value := range entry.Data {
		switch value := value.(type) {
		case error:

			errorMap["message"] = value.Error()
			gelfEntry["error"] = errorMap
		default:
			if !protectedFields[key] {
				contextMap[key] = value
			}
		}
	}

	if len(contextMap) > 0 {
		gelfEntry["contextMap"] = contextMap
	}
	message, err := json.Marshal(gelfEntry)
	return append(message, '\n'), err
}

func toTimestamp(t time.Time) float64 {
	nanosecond := float64(t.Nanosecond()) / 1e9
	seconds := float64(t.Unix())
	return seconds + nanosecond
}

func toEpochUnixTime(t time.Time) int64 {
	return t.Unix()
}

func toSyslogLevel(level logrus.Level) syslog.Priority {
	syslog, ok := levelMap[level]
	if ok {
		return syslog
	}
	return DefaultLevel
}
