package logging

import (
	"github.com/Graylog2/go-gelf/gelf"
	"github.com/bytetwiddler/digger/config"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

type asyncWriter struct {
	ch chan []byte
}

func (w *asyncWriter) Write(p []byte) (n int, err error) {
	msg := make([]byte, len(p))
	copy(msg, p)
	w.ch <- msg
	return len(p), nil
}

func SetupLogging(config *config.Config) (chan []byte, error) {
	// Set up GELF logger
	gelfWriter, err := gelf.NewWriter(config.Log.Gelf.Address)
	if err != nil {
		return nil, err
	}

	// Set up file logger with rotation
	fileLogger := &lumberjack.Logger{
		Filename:   config.Log.File.Filename,
		MaxSize:    config.Log.File.MaxSize,    // megabytes
		MaxBackups: config.Log.File.MaxBackups, // number of backups
		MaxAge:     config.Log.File.MaxAge,     // days
	}

	// Set log level
	level, err := logrus.ParseLevel(config.Log.Level)
	if err != nil {
		return nil, err
	}
	logrus.SetLevel(level)

	// Create a buffered channel for log messages
	logChannel := make(chan []byte, 100)

	// Start a goroutine to handle log messages asynchronously
	go func() {
		for msg := range logChannel {
			_, err := gelfWriter.Write(msg)
			if err != nil {
				logrus.Errorf("Failed to write to GELF: %v", err)
			}
			_, err = fileLogger.Write(msg)
			if err != nil {
				logrus.Errorf("Failed to write to file: %v", err)
			}
		}
	}()

	// Custom writer to send log messages to the channel
	logrus.SetOutput(&asyncWriter{logChannel})

	return logChannel, nil
}
