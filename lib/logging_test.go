package mail2most

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogging(t *testing.T) {
	m2m, err := New("../conf/mail2most.conf")
	assert.Nil(t, err)

	m2m.Error("Error - test", map[string]interface{}{})
	m2m.Info("Info - test", map[string]interface{}{})
	m2m.Debug("Debug - test", map[string]interface{}{})

	m2m.Config.Logging.Loglevel = "debug"
	m2m.Debug("Debug - test", map[string]interface{}{})
}
