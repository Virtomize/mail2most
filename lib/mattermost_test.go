package mail2most

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMattermost(t *testing.T) {
	m2m, err := New("../conf/mail2most.conf")
	assert.Nil(t, err)

	_, err = m2m.mlogin(0)
	assert.NotNil(t, err)

	err = m2m.PostMattermost(0, Mail{})
	assert.NotNil(t, err)
}
