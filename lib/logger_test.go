package mail2most

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	var conf config
	err := parseConfig("../conf/mail2most.conf", &conf)
	assert.Nil(t, err)

	m2m := Mail2Most{Config: conf}

	err = m2m.initLogger()
	assert.Nil(t, err)

	conf.Logging.Logtype = "json"
	conf.Logging.Loglevel = "debug"
	conf.Logging.Output = "foo"
	m2m = Mail2Most{Config: conf}
	err = m2m.initLogger()
	assert.Nil(t, err)

	conf.Logging.Logtype = "bar"
	conf.Logging.Loglevel = "error"
	conf.Logging.Output = "logfile"
	conf.Logging.Logfile = "/tmp/doesnotexists/mail2most.log"
	m2m = Mail2Most{Config: conf}
	err = m2m.initLogger()
	assert.NotNil(t, err)
	if err != nil {
		assert.Equal(t, err.Error(), "Can't open logfile: /tmp/doesnotexists/mail2most.log")
	}

	conf.Logging.Logfile = "/tmp/mail2most.log"
	m2m = Mail2Most{Config: conf}
	err = m2m.initLogger()
	assert.Nil(t, err)
	os.Remove("/tmp/mail2most.log")
}
