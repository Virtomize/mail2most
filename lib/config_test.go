package mail2most

import (
	"testing"

	filet "github.com/Flaque/filet"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfigNoFile(t *testing.T) {
	var conf config
	err := parseConfig("some Random file", &conf)
	assert.NotNil(t, err)
}

func TestLoadConfigEmptyConfig(t *testing.T) {
	defer filet.CleanUp(t)
	var conf config

	file := filet.TmpFile(t, "", "")

	err := parseConfig(file.Name(), &conf)
	assert.Nil(t, err)
}

func TestLoadConfigNil(t *testing.T) {
	defer filet.CleanUp(t)

	file := filet.TmpFile(t, "", "")

	err := parseConfig(file.Name(), nil)
	assert.NotNil(t, err)
}
