package logging

import (
	"testing"

	"github.com/Graylog2/go-gelf/gelf"
	"github.com/bytetwiddler/digger/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSetupLogging(t *testing.T) {
	cfg := &config.Config{
		Log: config.LogConfig{
			Level: "debug",
			GELF: config.GELFConfig{
				Address: "127.0.0.1:12201",
			},
			LogFile: config.LogFileConfig{
				Filename:   "test.log",
				MaxSize:    5,
				MaxBackups: 2,
				MaxAge:     15,
			},
		},
	}

	logChannel, err := SetupLogging(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, logChannel)

	// Close the log channel to clean up
	close(logChannel)
}

func TestSetupLoggingInvalidGELFAddress(t *testing.T) {
	cfg := &config.Config{
		Log: config.LogConfig{
			Level: "debug",
			GELF: config.GELFConfig{
				Address: "invalid_address",
			},
			LogFile: config.LogFileConfig{
				Filename:   "test.log",
				MaxSize:    5,
				MaxBackups: 2,
				MaxAge:     15,
			},
		},
	}

	logChannel, err := SetupLogging(cfg)
	assert.Error(t, err)
	assert.Nil(t, logChannel)
}

func TestGelfHookFire(t *testing.T) {
	writer, err := gelf.NewWriter("127.0.0.1:12201")
	assert.NoError(t, err)
	assert.NotNil(t, writer)

	hook := &GelfHook{writer: writer}

	entry := &logrus.Entry{
		Message: "test message",
		Level:   logrus.InfoLevel,
		Logger:  logrus.New(), // Ensure Logger is initialized
	}

	err = hook.Fire(entry)
	assert.NoError(t, err)
}

func TestGelfHookFireDifferentLevels(t *testing.T) {
	writer, err := gelf.NewWriter("127.0.0.1:12201")
	assert.NoError(t, err)
	assert.NotNil(t, writer)

	hook := &GelfHook{writer: writer}

	levels := []logrus.Level{
		logrus.TraceLevel,
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}

	for _, level := range levels {
		entry := &logrus.Entry{
			Message: "test message",
			Level:   level,
			Logger:  logrus.New(), // Ensure Logger is initialized
		}

		err = hook.Fire(entry)
		assert.NoError(t, err)
	}
}
