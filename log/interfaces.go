package log

import "time"

const (
	LOG_FIELD_TYPE_STRING = iota
	LOG_FIELD_TYPE_INTEGER
	LOG_FIELD_TYPE_BOOLEAN
	LOG_FIELD_TYPE_FLOAT
	LOG_FIELD_TYPE_OBJECT
	LOG_FIELD_TYPE_ERROR
	LOG_FIELD_TYPE_ARRAY
	LOG_FIELD_TYPE_TIME
	LOG_FIELD_TYPE_DURATION
)

type LogField struct {
	Key       string
	Type      uint8
	Integer   int64
	Float     float64
	String    string
	Interface interface{}
}

func LogError(err error) LogField {
	return LogField{
		Key:       "error",
		Type:      LOG_FIELD_TYPE_ERROR,
		Interface: err,
	}
}

func LogErrorWithKey(key string, err error) LogField {
	return LogField{
		Key:       key,
		Type:      LOG_FIELD_TYPE_ERROR,
		Interface: err,
	}
}

func LogString(key string, value string) LogField {
	return LogField{
		Key:    key,
		Type:   LOG_FIELD_TYPE_STRING,
		String: value,
	}
}
func LogInteger(key string, value int64) LogField {
	return LogField{
		Key:     key,
		Type:    LOG_FIELD_TYPE_INTEGER,
		Integer: value,
	}
}
func LogInt(key string, value int) LogField {
	return LogField{
		Key:     key,
		Type:    LOG_FIELD_TYPE_INTEGER,
		Integer: int64(value),
	}
}
func LogFloat(key string, value float64) LogField {
	return LogField{
		Key:   key,
		Type:  LOG_FIELD_TYPE_FLOAT,
		Float: value,
	}
}
func LogBoolean(key string, value bool) LogField {
	if value {
		return LogField{
			Key:     key,
			Type:    LOG_FIELD_TYPE_BOOLEAN,
			Integer: 1,
		}
	}
	return LogField{
		Key:     key,
		Type:    LOG_FIELD_TYPE_BOOLEAN,
		Integer: 0,
	}
}
func LogBool(key string, value bool) LogField {
	if value {
		return LogField{
			Key:     key,
			Type:    LOG_FIELD_TYPE_BOOLEAN,
			Integer: 1,
		}
	}
	return LogField{
		Key:     key,
		Type:    LOG_FIELD_TYPE_BOOLEAN,
		Integer: 0,
	}
}
func LogObject(key string, value interface{}) LogField {
	return LogField{
		Key:       key,
		Type:      LOG_FIELD_TYPE_OBJECT,
		Interface: value,
	}
}
func LogArray(key string, value []interface{}) LogField {
	return LogField{
		Key:       key,
		Type:      LOG_FIELD_TYPE_ARRAY,
		Interface: value,
	}
}
func LogTime(key string, value time.Time) LogField {
	return LogField{
		Key:       key,
		Type:      LOG_FIELD_TYPE_TIME,
		Interface: value,
	}
}
func LogDuration(key string, value time.Duration) LogField {
	return LogField{
		Key:       key,
		Type:      LOG_FIELD_TYPE_DURATION,
		Interface: value,
	}
}

// ILogger is an interface that defines the methods
// for logging messages at different levels.
// It is used to provide a consistent logging interface
// across different implementations of logging libraries.
type ILogger interface {
	// Info logs a message at the Info level.
	Info(message string, fields ...LogField)
	// Debug logs a message at the Debug level.
	Debug(message string, fields ...LogField)
	// Warn logs a message at the Warn level.
	Warn(message string, fields ...LogField)
	// Error logs a message at the Error level.
	Error(message string, fields ...LogField)
	// Fatal logs a message at the Fatal level.
	// This level is used for critical errors that cause the application to terminate.
	// The application will exit after logging the message.
	Fatal(message string, fields ...LogField)
	// Sync flushes any buffered log entries.
	Sync() error
}
