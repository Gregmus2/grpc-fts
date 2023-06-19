package config

import "github.com/sirupsen/logrus"

func NewLogrusEntry() (*logrus.Entry, error) {
	logger := logrus.New()

	logger.SetFormatter(&logrus.JSONFormatter{
		DisableTimestamp: true,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "severity",
		},
	})

	logger.SetLevel(logrus.InfoLevel)

	return logrus.NewEntry(logger), nil
}
