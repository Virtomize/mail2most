package mail2most

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type htmlTest struct {
	Body []byte
	Err  error
	Res  []byte
}

func TestParseHTML(t *testing.T) {
	m2m, err := New("../conf/mail2most.conf")
	assert := assert.New(t)
	assert.Nil(err)

	// ignore test
	body := []byte("An error occurred while trying to deliver the mail to the following recipients:")
	b, err := m2m.parseHTML(body, 0)
	assert.NotNil(err)
	assert.Equal(b, []byte{})

	m2m.Config.Profiles[0].Filter.IgnoreMailErrorNotifications = false
	b, err = m2m.parseHTML(body, 0)
	assert.Nil(err)
	assert.Equal(b, body)
}
