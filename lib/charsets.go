package mail2most

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

var charsets = map[string]encoding.Encoding{
	"windows-874":  charmap.Windows1250,
	"windows-1253": charmap.Windows1253,
	"windows-1254": charmap.Windows1254,
	"windows-1255": charmap.Windows1255,
	"windows-1256": charmap.Windows1256,
	"windows-1257": charmap.Windows1257,
	"windows-1258": charmap.Windows1258,
}
