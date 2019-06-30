package mail2most

// Error logging
func (m Mail2Most) Error(msg string, params map[string]interface{}) {
	m.Logger.WithFields(params).Error(msg)
}

// Info logging
func (m Mail2Most) Info(msg string, params map[string]interface{}) {
	if m.Config.Logging.Loglevel == INFO || m.Config.Logging.Loglevel == DEBUG {
		m.Logger.WithFields(params).Info(msg)
	}
}

// Debug logging
func (m Mail2Most) Debug(msg string, params map[string]interface{}) {
	if m.Config.Logging.Loglevel == DEBUG {
		m.Logger.WithFields(params).Debug(msg)
	}
}
