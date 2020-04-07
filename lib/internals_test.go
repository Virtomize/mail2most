package mail2most

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	imap "github.com/emersion/go-imap"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	_, err := New("doesnotexists")
	assert.NotNil(t, err)
	if err != nil {
		assert.Equal(t, err.Error(), "stat doesnotexists: no such file or directory")
	}

	_, err = New("../conf/testing.conf")
	assert.NotNil(t, err)
	if err != nil {
		assert.Equal(t, err.Error(), "Can't open logfile: /tmp/doesnotexists/mail2most.log")
	}

	_, err = New("../conf/mail2most.conf")
	assert.Nil(t, err)
}

func TestFilters(t *testing.T) {
	m2m, err := New("../conf/mail2most.conf")
	assert.Nil(t, err)

	mail := Mail{
		From:    []*imap.Address{&imap.Address{MailboxName: "test", HostName: "example.com"}},
		To:      []*imap.Address{&imap.Address{MailboxName: "info", HostName: "example.com"}},
		Subject: "i am an example subject",
		Date:    time.Now(),
	}
	test := m2m.containsFrom(0, mail)
	assert.True(t, test)

	mail.From[0].MailboxName = "test2"
	test = m2m.containsFrom(0, mail)
	assert.False(t, test)

	test = m2m.containsTo(0, mail)
	assert.True(t, test)

	mail.To[0].MailboxName = "info2"
	test = m2m.containsTo(0, mail)
	assert.False(t, test)

	mail.To = []*imap.Address{
		&imap.Address{MailboxName: "info", HostName: "example.de"},
		&imap.Address{MailboxName: "info", HostName: "example.com"},
	}
	test = m2m.containsTo(0, mail)
	assert.True(t, test)

	test = m2m.containsSubject(0, mail)
	assert.True(t, test)

	mail.Subject = "foo"
	test = m2m.containsSubject(0, mail)
	assert.False(t, test)

	test, err = m2m.filterByTimeRange(0, mail)
	assert.Nil(t, err)
	assert.True(t, test)

	mail.Date = time.Now().AddDate(-1, 0, 0)
	test, err = m2m.filterByTimeRange(0, mail)
	assert.Nil(t, err)
	assert.False(t, test)

	mail.Date = time.Now()
	m2m.Config.Profiles[0].Filter.TimeRange = "foo"
	_, err = m2m.filterByTimeRange(0, mail)
	assert.NotNil(t, err)
	if err != nil {
		assert.Equal(t, err.Error(), "time: invalid duration foo")
	}

	mail.Subject = "i am an example subject"
	m2m.Config.Profiles[0].Filter.TimeRange = "24h"
	mail.From[0].MailboxName = "test"
	test, err = m2m.checkFilters(0, mail)
	assert.Nil(t, err)
	assert.True(t, test)

	mail.From[0].MailboxName = "test2"
	test, err = m2m.checkFilters(0, mail)
	assert.Nil(t, err)
	assert.False(t, test)

	mail.From[0].MailboxName = "test"
	mail.To[1].MailboxName = "foo"
	test, err = m2m.checkFilters(0, mail)
	assert.Nil(t, err)
	assert.False(t, test)

	mail.To[1].MailboxName = "info"
	mail.Date = time.Now().AddDate(-1, 0, 0)
	test, err = m2m.checkFilters(0, mail)
	assert.Nil(t, err)
	assert.False(t, test)

	mail.To[1].MailboxName = "info"
	m2m.Config.Profiles[0].Filter.TimeRange = "foo"
	_, err = m2m.checkFilters(0, mail)
	assert.NotNil(t, err)
	if err != nil {
		assert.Equal(t, err.Error(), "time: invalid duration foo")
	}

	// empty filters
	m2m.Config.Profiles[0].Filter.From = []string{}
	m2m.Config.Profiles[0].Filter.To = []string{}
	m2m.Config.Profiles[0].Filter.Subject = []string{}
	m2m.Config.Profiles[0].Filter.TimeRange = ""

	test = m2m.containsFrom(0, mail)
	assert.True(t, test)

	test = m2m.containsTo(0, mail)
	assert.True(t, test)

	test = m2m.containsSubject(0, mail)
	assert.True(t, test)

	test, err = m2m.filterByTimeRange(0, mail)
	assert.Nil(t, err)
	assert.True(t, test)

}

func TestWriteToFile(t *testing.T) {
	data := make([][]uint32, 1)
	err := writeToFile(data, "/tmp/delete.me")
	assert.Nil(t, err)

	os.Remove("/tmp/delete.me")

	err = writeToFile(data, "/tmp/doesnotexists/delete.me")
	assert.NotNil(t, err)
	if err != nil {
		assert.Equal(t, err.Error(), "open /tmp/doesnotexists/delete.me: no such file or directory")
	}
}

func TestRead(t *testing.T) {
	m2m, err := New("../conf/mail2most.conf")
	assert.Nil(t, err)

	_, err = m2m.read(nil)
	assert.Equal(t, err, fmt.Errorf("nil reader"))

	_, err = m2m.read(strings.NewReader(strings.ReplaceAll(testMailString, "windows-1252", "asdf")))
	assert.Nil(t, err)

	mr, err := m2m.read(strings.NewReader(testMailString))
	assert.Nil(t, err)

	_, _, err = m2m.processReader(nil, 0)
	assert.Equal(t, err, fmt.Errorf("nil reader"))

	b, _, err := m2m.processReader(mr, 1)
	assert.Nil(t, err)
	assert.Equal(t, b, "What's <i>your</i> name?")
}
