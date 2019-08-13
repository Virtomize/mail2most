package mail2most

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	m2m, err := New("../conf/mail2most.conf")
	assert.Nil(t, err)

	_, err = os.Create("/tmp/data.json")
	assert.Nil(t, err)

	m2m.Config.General.File = "/tmp/data.json"
	err = m2m.Run()
	assert.NotNil(t, err)
	if err != nil {
		assert.Equal(t, err.Error(), "unexpected end of JSON input")
	}

	os.Chmod("/tmp/data.json", 0200)
	err = m2m.Run()
	assert.NotNil(t, err)
	if err != nil {
		assert.Equal(t, err.Error(), "open /tmp/data.json: permission denied")
	}

	err = os.Remove("/tmp/data.json")
	assert.Nil(t, err)
}
