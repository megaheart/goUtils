package log

import (
	"os"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger_Output defines where the logger will write output.
// It is used to select between console, file, or both outputs.
type ZapLogger_Output int

const (
	// zero value is unused
	_ ZapLogger_Output = iota
	// ZapLogger_Output_Console writes logs to stdout.
	ZapLogger_Output_Console
	// ZapLogger_Output_File writes logs to a file.
	ZapLogger_Output_File
	// ZapLogger_Output_ConsoleAndFile writes logs to both stdout and a file.
	ZapLogger_Output_ConsoleAndFile
)

// ZapLogger_Format selects the encoding format for log messages.
// Use JSON for structured logs or ReadableText for console-friendly output.
type ZapLogger_Format int

const (
	// zero value is unused
	_ ZapLogger_Format = iota
	// ZapLogger_Format_Json encodes logs as JSON objects.
	ZapLogger_Format_Json
	// ZapLogger_Format_ReadableText encodes logs in a colored, human-friendly text format.
	ZapLogger_Format_ReadableText
)

type ZapLogger_Level int

const (
	ZapLogger_Level_Debug ZapLogger_Level = iota - 1
	ZapLogger_Level_Info
	ZapLogger_Level_Warn
	ZapLogger_Level_Error
	ZapLogger_Level_DPanic
	ZapLogger_Level_Panic
	ZapLogger_Level_Fatal

	ZapLogger_Level_Min = ZapLogger_Level_Debug
	ZapLogger_Level_Max = ZapLogger_Level_Fatal

	ZapLogger_Level_Invalid = ZapLogger_Level_Max + 1
)

func (l ZapLogger_Level) ToZapLevel() zapcore.Level {
	return zapcore.Level(l) // Map Debug=0 to zapcore.DebugLevel=-1, Info=1 to zapcore.InfoLevel=0, etc.
}

// ZapcoreSamplerOption configures zapcore's sampling behavior.
//
// This struct maps directly to the parameters accepted by
// zapcore.NewSamplerWithOptions and controls how frequently identical
// or high-volume log messages are actually emitted. Use sampling to
// reduce noise and disk/network usage for hot loops or repeated errors.
//
// Field details:
//
//   - Tick: sampling window duration. Each Tick interval the sampler
//     resets its counters. Typical values are in the order of
//     hundreds of milliseconds up to a few seconds (e.g. 1s).
//
//   - First: the number of log events allowed unconditionally at the
//     start of each Tick. For example, First=5 means the first 5
//     messages in each Tick are always logged (no sampling).
//
//   - Thereafter: after the initial First events in a Tick, the sampler
//     will allow 1 event every Thereafter occurrences. For example,
//     Thereafter=10 permits one log out of every 10 events after the
//     first First events. If Thereafter <= 0, sampling after First is
//     effectively disabled (only the First events are guaranteed).
//
//   - Options: additional functional options of type zapcore.SamplerOption
//     that modify sampler behavior (these are passed through to
//     zapcore.NewSamplerWithOptions). This can usually be left nil,
//     but advanced callers can supply zap-provided options if needed.
//
// Example: to allow the first 5 events every second and then sample
// 1/10 of subsequent events, use Tick=1*time.Second, First=5,
// Thereafter=10.
type ZapcoreSamplerOption struct {
	// Tick is the sampling interval that resets the sampler counters.
	Tick time.Duration

	// First is how many events in each Tick are logged without sampling.
	First int

	// Thereafter controls how often to sample after the initial First events
	// (1 event allowed every `Thereafter` occurrences). If <= 0, no sampling
	// occurs after the First events.
	Thereafter int

	// Options holds optional zapcore.SamplerOption values passed to
	// zapcore.NewSamplerWithOptions for fine-grained sampler configuration.
	Options []zapcore.SamplerOption
}

// NewZapLogger constructs an ILogger backed by Uber's zap logger.
// Parameters:
//   - format: choose JSON or ReadableText encoding
//   - timeLayout: time layout used when formatting timestamps for readable text
//   - outputType: where to write logs (console, file, or both)
//   - logPath: path to the log file when using file output; if empty, console is used
//   - level: zapcore level to filter logs (e.g. zapcore.InfoLevel)
//   - samplerOption: optional sampling configuration to reduce high-frequency log noise
func NewZapLogger(
	format ZapLogger_Format,
	timeLayout string,
	outputType ZapLogger_Output,
	logPath string,
	level ZapLogger_Level,
	samplerOption *ZapcoreSamplerOption,
) ILogger {
	var cores []zapcore.Core

	switch outputType {
	case ZapLogger_Output_Console:
		// console: encoder chosen according to the selected format
		var enc zapcore.Encoder
		if format == ZapLogger_Format_ReadableText {
			enc = buildConsoleEncoder(timeLayout, true) // colored
		} else {
			enc = buildJSONEncoder()
		}
		cores = append(cores, zapcore.NewCore(enc, zapcore.AddSync(os.Stdout), level.ToZapLevel()))

	case ZapLogger_Output_File:
		// file: plain output (no color) — JSON is recommended
		fileWS := fileSink(logPath)
		var enc zapcore.Encoder
		if format == ZapLogger_Format_ReadableText {
			enc = buildConsoleEncoder(timeLayout, false) // non-colored
		} else {
			enc = buildJSONEncoder()
		}
		cores = append(cores, zapcore.NewCore(enc, fileWS, level.ToZapLevel()))

	case ZapLogger_Output_ConsoleAndFile:
		// console (colored) + file (non-colored / JSON)
		consoleEnc := buildConsoleEncoder(timeLayout, true)
		consoleWS := zapcore.AddSync(os.Stdout)

		var fileEnc zapcore.Encoder
		if format == ZapLogger_Format_ReadableText {
			fileEnc = buildConsoleEncoder(timeLayout, false) // non-colored
		} else {
			fileEnc = buildJSONEncoder()
		}
		fileWS := fileSink(logPath)

		cores = append(cores,
			zapcore.NewCore(consoleEnc, consoleWS, level.ToZapLevel()),
			zapcore.NewCore(fileEnc, fileWS, level.ToZapLevel()),
		)

	default:
		// fallback: colored console
		enc := buildConsoleEncoder(timeLayout, true)
		cores = append(cores, zapcore.NewCore(enc, zapcore.AddSync(os.Stdout), level.ToZapLevel()))
	}

	// combine cores into a single core
	core := zapcore.NewTee(cores...)

	// Sampler (optional)
	if samplerOption != nil {
		core = zapcore.NewSamplerWithOptions(core, samplerOption.Tick, samplerOption.First, samplerOption.Thereafter, samplerOption.Options...)
	}

	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.AddCallerSkip(1),
	)

	return &ZapLogger{logger: logger}
}

func buildConsoleEncoder(timeLayout string, colored bool) zapcore.Encoder {
	encCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.TimeEncoderOfLayout(timeLayout),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	if colored {
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		encCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	return zapcore.NewConsoleEncoder(encCfg)
}

func buildJSONEncoder() zapcore.Encoder {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.TimeKey = "ts"
	return zapcore.NewJSONEncoder(cfg)
}

// fileSink returns a WriteSyncer that writes to a rotating file using lumberjack.
func fileSink(path string) zapcore.WriteSyncer {
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   path,
		MaxSize:    100, // MB
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	})
}

// ZapLogger is a thin wrapper around *zap.Logger implementing the ILogger interface.
// It converts the package's LogField types into zap fields and forwards calls.
type ZapLogger struct {
	logger *zap.Logger
	// stopProgram func()
}

// convertFieldsToZap maps the package LogField representation into zap.Field values.
// This function handles various typed fields (string, int, float, bool, object, array, time, duration).
func convertFieldsToZap(fields []LogField) []zap.Field {
	var zapFields []zap.Field
	for _, field := range fields {
		switch field.Type {
		case LOG_FIELD_TYPE_STRING:
			zapFields = append(zapFields, zap.String(field.Key, field.String))
		case LOG_FIELD_TYPE_INTEGER:
			zapFields = append(zapFields, zap.Int64(field.Key, field.Integer))
		case LOG_FIELD_TYPE_FLOAT:
			zapFields = append(zapFields, zap.Float64(field.Key, field.Float))
		case LOG_FIELD_TYPE_BOOLEAN:
			zapFields = append(zapFields, zap.Bool(field.Key, field.Integer == 1))
		case LOG_FIELD_TYPE_OBJECT:
			zapFields = append(zapFields, zap.Object(field.Key, field.Interface.(zapcore.ObjectMarshaler)))
		case LOG_FIELD_TYPE_ARRAY:
			zapFields = append(zapFields, zap.Array(field.Key, field.Interface.(zapcore.ArrayMarshaler)))
		case LOG_FIELD_TYPE_TIME:
			zapFields = append(zapFields, zap.Time(field.Key, field.Interface.(time.Time)))
		case LOG_FIELD_TYPE_DURATION:
			zapFields = append(zapFields, zap.Duration(field.Key, field.Interface.(time.Duration)))
		default:
			zapFields = append(zapFields, zap.Any(field.Key, field.Interface))
		}
	}
	return zapFields
}

// Info logs an info-level message with optional structured fields.
func (l *ZapLogger) Info(message string, fields ...LogField) {
	l.logger.Info(message, convertFieldsToZap(fields)...)
}

// Debug logs a debug-level message with optional structured fields.
func (l *ZapLogger) Debug(message string, fields ...LogField) {
	l.logger.Debug(message, convertFieldsToZap(fields)...)
}

// Warn logs a warning-level message with optional structured fields.
func (l *ZapLogger) Warn(message string, fields ...LogField) {
	l.logger.Warn(message, convertFieldsToZap(fields)...)
}

// Error logs an error-level message with optional structured fields.
func (l *ZapLogger) Error(message string, fields ...LogField) {
	l.logger.Error(message, convertFieldsToZap(fields)...)
}

// Fatal logs a fatal-level message and then terminates the program.
func (l *ZapLogger) Fatal(message string, fields ...LogField) {
	l.logger.Fatal(message, convertFieldsToZap(fields)...)
}

// Sync flushes any buffered log entries.
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}
