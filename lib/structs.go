package mail2most

import (
	"time"

	imap "github.com/emersion/go-imap"
	log "github.com/sirupsen/logrus"
)

// Mail2Most implements the basic interface
type Mail2Most struct {
	Config config
	Logger *log.Logger
}

// Mail contains mail information
type Mail struct {
	ID            uint32
	Subject, Body string
	From, To      []*imap.Address
	Date          time.Time
	Attachments   []Attachment
}

// Attachment .
type Attachment struct {
	Filename string
	Content  []byte
}
