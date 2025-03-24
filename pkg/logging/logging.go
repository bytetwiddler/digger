package logging

import (
	"fmt"

	"github.com/Graylog2/go-gelf/gelf"
	"github.com/bytetwiddler/digger/pkg/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// SetupLogging configures the logging for the application.
func SetupLogging(cfg *config.Config) (chan struct{}, error) {
	logrus.SetLevel(parseLogLevel(cfg.Log.Level))

	// Setup file logging with rotation
	logrus.SetOutput(&lumberjack.Logger{
		Filename:   cfg.Log.LogFile.Filename,
		MaxSize:    cfg.Log.LogFile.MaxSize,    // megabytes
		MaxBackups: cfg.Log.LogFile.MaxBackups, // number of backups
		MaxAge:     cfg.Log.LogFile.MaxAge,     // days
	})

	// Setup Graylog logging
	gelfWriter, err := gelf.NewWriter(cfg.Log.GELF.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to create GELF writer: %w", err)
	}

	logrus.AddHook(&GelfHook{writer: gelfWriter})

	// Create a channel to signal when to close the loggers
	logChannel := make(chan struct{})

	go func() {
		<-logChannel
		_ = gelfWriter.Close()
	}()

	return logChannel, nil
}

func parseLogLevel(level string) logrus.Level {
	switch level {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

// GelfHook is a logrus hook for sending logs to Graylog.
type GelfHook struct {
	writer *gelf.Writer
}

// Levels returns the log levels supported by the hook.
func (hook *GelfHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire sends the log entry to Graylog.
func (hook *GelfHook) Fire(entry *logrus.Entry) error {
	if entry == nil {
		return nil // or return an appropriate error
	}

	line, err := entry.String()
	if err != nil {
		return fmt.Errorf("failed to convert log entry to string: %w", err)
	}

	level := int32(entry.Level)
	if level < 0 || level > 2147483647 {
		return fmt.Errorf("log level out of range: %d", level)
	}

	return hook.writer.WriteMessage(&gelf.Message{
		Version:  "1.1",
		Host:     "localhost", // Replace with actual host if needed
		Short:    line,
		Full:     line,
		TimeUnix: float64(entry.Time.UnixNano()) / 1e9,
		Level:    level,
	})
}
