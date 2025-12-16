package logs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	MessageZapKey  string = "message"
	SeverityZapKey string = "severity"
	TimeZapKey     string = "time"
)

type StructuredLogger interface {
	Infof(format string, v ...any)
	Warnf(format string, v ...any)
	Errorf(format string, v ...any)
	Infow(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
	Println(v ...any)
}

type ZapStructuredLogger struct {
	logger *zap.SugaredLogger
}

func severityEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var severity string
	switch level {
	case zapcore.ErrorLevel:
		severity = "ERROR"
	case zapcore.WarnLevel:
		severity = "WARNING"
	case zapcore.InfoLevel:
		severity = "INFO"
	case zapcore.DebugLevel:
		severity = "DEBUG"
	default:
		severity = "DEFAULT"
	}
	enc.AppendString(severity)
}

func newZapLogger(file string) (*ZapStructuredLogger, error) {
	cfg := zap.NewProductionConfig()
	cfg.DisableCaller = true
	cfg.DisableStacktrace = true
	cfg.EncoderConfig.MessageKey = MessageZapKey
	cfg.EncoderConfig.LevelKey = SeverityZapKey
	cfg.EncoderConfig.TimeKey = TimeZapKey
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	cfg.EncoderConfig.EncodeLevel = severityEncoder
	cfg.OutputPaths = []string{file}
	logger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	return &ZapStructuredLogger{logger: logger.Sugar()}, nil
}

func NewFileLogger(path, prefix string) (*ZapStructuredLogger, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("could not create directory: %w", err)
	}
	fileId, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("could not create UUID: %w", err)
	}
	fileName := fmt.Sprintf("%s-%s.json", prefix, fileId.String())
	filePath := filepath.Join(path, fileName)
	fmt.Printf("writing to: %s\n", filePath)
	return newZapLogger(filePath)
}

func (f ZapStructuredLogger) Infof(format string, v ...any) {
	f.logger.Infof(format, v...)
}
func (f ZapStructuredLogger) Warnf(format string, v ...any) {
	f.logger.Warnf(format, v...)
}
func (f ZapStructuredLogger) Errorf(format string, v ...any) {
	f.logger.Errorf(format, v...)
}
func (f ZapStructuredLogger) Infow(msg string, keysAndValues ...any) {
	f.logger.Infow(msg, keysAndValues...)
}
func (f ZapStructuredLogger) Warnw(msg string, keysAndValues ...any) {
	f.logger.Warnw(msg, keysAndValues...)
}
func (f ZapStructuredLogger) Errorw(msg string, keysAndValues ...any) {
	f.logger.Errorw(msg, keysAndValues...)
}
func (f ZapStructuredLogger) Println(v ...any) {
	f.logger.Infoln(v...)
}

type SimpleLogger struct {
	l *log.Logger
}

func (sl SimpleLogger) Infof(format string, v ...any) {
	sl.l.Printf(format, v...)
}
func (sl SimpleLogger) Warnf(format string, v ...any) {
	sl.l.Printf(format, v...)
}
func (sl SimpleLogger) Errorf(format string, v ...any) {
	sl.l.Printf(format, v...)
}
func (sl SimpleLogger) Infow(msg string, keysAndValues ...any) {
	fields := map[string]any{"message": msg}
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			if key, ok := keysAndValues[i].(string); ok {
				fields[key] = keysAndValues[i+1]
			}
		}
	}
	b, err := json.Marshal(fields)
	if err != nil {
		sl.l.Printf("error marshaling log: %v", err)
		return
	}
	sl.l.Println(string(b))
}
func (sl SimpleLogger) Warnw(msg string, keysAndValues ...any) {
	sl.Infow(msg, keysAndValues...)
}
func (sl SimpleLogger) Errorw(msg string, keysAndValues ...any) {
	sl.Infow(msg, keysAndValues...)
}
func (sl SimpleLogger) Println(v ...any) {
	sl.l.Println(v...)
}

func NewStdoutLogger() SimpleLogger {
	return SimpleLogger{log.New(os.Stdout, "", 0)}
}
