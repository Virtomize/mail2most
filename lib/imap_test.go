package mail2most

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMail(t *testing.T) {
	m2m, err := New("../conf/mail2most.conf")
	assert.Nil(t, err)

	_, err = m2m.connect(0)
	assert.NotNil(t, err)

	_, err = m2m.GetMail(0)
	assert.NotNil(t, err)

	_, err = m2m.ListMailBoxes(0)
	assert.NotNil(t, err)

	_, err = m2m.ListFlags(0)
	assert.NotNil(t, err)
}
