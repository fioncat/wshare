package log

import (
	"fmt"
	"strings"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/fioncat/wshare/config"
	"github.com/sirupsen/logrus"
)

var (
	logger *logrus.Logger

	initOnce sync.Once
)

// Init inits the log, MUST be called after `config.Init()`.
// MUST be called before using `log.Get()`
func Init() error {
	var err error
	initOnce.Do(func() {
		err = doInit(config.Get().Log)
	})
	return err
}

func doInit(cfg *config.Log) error {
	logger = logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   true,
	})

	switch cfg.Level {
	case "", "info":
		logger.SetLevel(logrus.InfoLevel)

	case "debug":
		logger.SetLevel(logrus.DebugLevel)

	case "error":
		logger.SetLevel(logrus.ErrorLevel)

	default:
		return fmt.Errorf("unknown log level: %q", cfg.Level)
	}

	return nil
}

func Get() *logrus.Logger {
	if logger == nil {
		panic("internal: please call log.Init before using Get()")
	}
	return logger
}

func BytesSize(data []byte) string {
	size := humanize.IBytes(uint64(len(data)))
	return strings.Replace(size, " ", "", 1)
}
