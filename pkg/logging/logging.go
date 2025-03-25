package logging

import (
	"fmt"
	"os"

	"github.com/Graylog2/go-gelf/gelf"
	"github.com/bytetwiddler/digger/pkg/config"
	"github.com/sirupsen/logrus"
)

type GelfHook struct {
	Writer *gelf.Writer
}

func (hook *GelfHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *GelfHook) Fire(entry *logrus.Entry) error {
	msg, err := entry.String()
	if err != nil {
		return fmt.Errorf("failed to setup GELF hook: %w", err)
	}

	writeErr := hook.Writer.WriteMessage(&gelf.Message{
		Version:  "1.1",
		Host:     "localhost",
		Short:    entry.Message,
		Full:     msg,
		TimeUnix: float64(entry.Time.UnixNano()) / 1e9, //nolint:mnd // magic is fine here
		Level:    int32(entry.Level),                   //nolint:gosec // This is const log level, not a user input
	})

	if writeErr != nil {
		return fmt.Errorf("failed to setup GELF: %w", writeErr)
	}

	return nil
}

func SetupLogging(cfg *config.Config) (*os.File, error) {
	// Set up logging to a file
	file, err := os.OpenFile(cfg.Log.File.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666) //nolint:mnd // magic is fine here
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	logrus.SetOutput(file)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	// Set log level
	level, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		return nil, fmt.Errorf("failed to set log level: %w", err)
	}

	logrus.SetLevel(level)

	// Set up logging to a GELF server
	gelfWriter, err := gelf.NewWriter(cfg.Log.Gelf.Address)

	if err != nil {
		return nil, fmt.Errorf("failed to setup GELF server: %w", err)
	}

	logrus.AddHook(&GelfHook{Writer: gelfWriter})

	return file, nil
}
