package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

func newLogger(out io.Writer) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(out)
	logger.SetLevel(logrus.WarnLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    true,
		DisableQuote:     true,
	})
	return logger
}

func configureLogger(logger *logrus.Logger, level string, disableColors bool) error {
	parsed, err := logrus.ParseLevel(strings.TrimSpace(level))
	if err != nil {
		return model.NewError(model.ErrorUsage, "invalid --log-level %q (expected panic, fatal, error, warn, info, debug, or trace)", level)
	}
	logger.SetLevel(parsed)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    disableColors,
		DisableQuote:     true,
	})
	return nil
}

func logCommandError(logger *logrus.Logger, err error) {
	if logger != nil {
		logger.Error(fmt.Sprintf("%v", err))
	}
}
