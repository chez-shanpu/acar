package logging

import "github.com/sirupsen/logrus"

type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"

	DefaultLogFormat = LogFormatText
	DefaultLogLevel  = logrus.InfoLevel
)

var DefaultLogger = InitializeDefaultLogger()

func InitializeDefaultLogger() (logger *logrus.Logger) {
	logger = logrus.New()
	logger.SetFormatter(GetFormatter(DefaultLogFormat))
	logger.SetLevel(DefaultLogLevel)
	return
}

func GetFormatter(format LogFormat) logrus.Formatter {
	switch format {
	case LogFormatText:
		return &logrus.TextFormatter{
			DisableTimestamp: true,
			//DisableColors:    true,
		}
	case LogFormatJSON:
		return &logrus.JSONFormatter{
			//DisableTimestamp: true,
		}
	}

	return nil
}
