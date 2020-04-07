package mail2most

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHTML(t *testing.T) {
	m2m, err := New("../conf/mail2most.conf")
	assert := assert.New(t)
	assert.Nil(err)

	// ignore test
	body := []byte("An error occurred while trying to deliver the mail to the following recipients:")
	b, err := m2m.parseHTML(body, 0)
	assert.NotNil(err)
	assert.Empty(b)

	m2m.Config.Profiles[0].Filter.IgnoreMailErrorNotifications = false
	b, err = m2m.parseHTML(body, 0)
	assert.Nil(err)
	assert.Equal(b, body)

	tests := []string{
		"<html></head>", // this works quite strange and should be refactored maybe
		`<div class="ms-outlook-ios-signature">
		foo`,
		"Sent with BlackBerry Work",
		`‐‐‐‐‐‐‐ Original Message ‐‐‐‐‐‐‐`,
		`On foo wrote: bar`,
		`Begin forwarded message:`,
		`&nbsp;`,
		`<style> foo </style>`,
		`<meta foo> bar </meta>`,
		`<div></div>`,
		`<o:p foo></o:p>`,
		`<span foo></span>`,
		`<img src="...">`,
		`<img src='...'>`,
		`<p></p>`,
		`<blockquote foo>`,
		`Sent from foo`,
		`Sent From foo`,
		`Sent via foo`,
	}

	for _, t := range tests {
		b, err := m2m.parseHTML([]byte(t), 0)
		assert.Nil(err)
		assert.Empty(b)
	}
}

func TestParseText(t *testing.T) {

	m2m, err := New("../conf/mail2most.conf")
	assert := assert.New(t)
	assert.Nil(err)

	tests := []string{
		`On foo wrote: 
	test`,
		"Begin forwarded message:",
	}

	for _, t := range tests {
		b, err := m2m.parseText([]byte(t))
		assert.Nil(err)
		assert.Empty(b)
	}
}
