package config

import "github.com/sirupsen/logrus"

func NewLogrusEntry(cfg *Global) (*logrus.Entry, error) {
	logger := logrus.New()
	var formatter logrus.Formatter

	switch cfg.Format {
	case "text":
		formatter = &logrus.TextFormatter{
			FullTimestamp: cfg.Timestamp,
		}
	case "json":
		formatter = &logrus.JSONFormatter{}
	case "stackdriver":
		formatter = &logrus.JSONFormatter{
			DisableTimestamp: !cfg.Timestamp,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyLevel: "severity",
			},
		}
	default:
		formatter = &logrus.TextFormatter{
			FullTimestamp: cfg.Timestamp,
		}
	}

	logger.SetFormatter(formatter)
	logger.SetLevel(logrus.InfoLevel)

	return logrus.NewEntry(logger), nil
}
