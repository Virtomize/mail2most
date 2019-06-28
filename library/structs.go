package mail2most

import imap "github.com/emersion/go-imap"

// Mail2Most implements the basic interface
type Mail2Most struct {
	Config config
}

// Mail contains mail information
type Mail struct {
	ID            uint32
	Subject, Body string
	From, To      []*imap.Address
}
