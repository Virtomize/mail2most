package mail2most

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func (m *Mail2Most) initLogger() error {

	m.Logger = log.StandardLogger()

	switch m.Config.Logging.Logtype {
	case LOGFORMATJSON:
		m.Logger.SetFormatter(&log.JSONFormatter{})
	case LOGFORMATTEXT:
		formatter := &log.TextFormatter{
			FullTimestamp: true,
		}
		m.Logger.SetFormatter(formatter)
	default:
		m.Logger.WithFields(log.Fields{
			"logformat": m.Config.Logging.Logtype,
			"default":   LOGFORMATTEXT,
		}).Error("unknown logformat using default")
	}

	switch m.Config.Logging.Loglevel {
	case INFO:
		m.Logger.SetLevel(log.InfoLevel)
	case ERROR:
		m.Logger.SetLevel(log.ErrorLevel)
	case DEBUG:
		m.Logger.SetLevel(log.DebugLevel)
	default:
		m.Logger.WithFields(log.Fields{
			"loglevel": m.Config.Logging.Loglevel,
			"default":  INFO,
		}).Error("unknown loglevel using default")
		log.SetLevel(log.InfoLevel)
	}

	switch m.Config.Logging.Output {
	case LOGFILE:
		logfile, err := os.OpenFile(m.Config.Logging.Logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			m.Logger.WithFields(log.Fields{
				"filepath": m.Config.Logging.Logfile,
			}).Error("can't open logfile")
			m.Config.Logging.Output = LOGSTDOUT
			return fmt.Errorf("Can't open logfile: %s", m.Config.Logging.Logfile)
		}
		m.Logger.SetOutput(logfile)
		m.Logger.WithFields(log.Fields{
			"output": LOGFILE,
			"format": m.Config.Logging.Logtype,
		}).Debug("initialising logging")
	case LOGSTDOUT:
		m.Logger.WithFields(log.Fields{
			"output": LOGSTDOUT,
			"format": m.Config.Logging.Logtype,
		}).Debug("using logging method")
	default:
		m.Logger.WithFields(log.Fields{
			"output":  m.Config.Logging.Output,
			"default": LOGSTDOUT,
		}).Error("unknown log output using default")
		m.Config.Logging.Output = LOGSTDOUT
	}
	return nil
}
