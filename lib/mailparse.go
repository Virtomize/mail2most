package mail2most

// credits go to
// https://github.com/justledbetter/mail2most/commit/b8e8aac8f472af6b39da161921afed5972f09c27
// thanks for the great work

//
// mail2most Mail Parser
//
// This extension to the mail2most app is intended to pull the "latest reply" off of the top of incoming
// e-mail messages, and only post that to the Mattermost channel. We do this to keep the Mattermost
// channel looking conversational. Otherwise, the user is spammed with entire copies of an e-mail
// conversation, including repetitive replies that may have already been posted to the channel (as they
// are included in the full text of the e-mail by default).
//

import (
	"crypto/sha256"
	"errors"
	"regexp"

	// image extensions
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// Track attachments that we have seen within an import session. Once we've posted an attachment to
// the channel once, we don't want to upload it again. This is to clean up channel text where users
// are in the habit of including the same attachments every time they hit Reply All in their mail
// client.
//
var seenAttachments map[[32]byte]string

// parseHtml attempts to strip everything out of the message body except for the latest reply. This is
// not perfect, but it's better than nothing. Different mail clients encode their message and replies
// in their own unique ways, and it's impossible to account for all of the potential variations.
//
// Returns the stripped message body, or null and an error.
//
func (m Mail2Most) parseHTML(b []byte, profile int) ([]byte, error) {

	// Is this an error message?  Nuke it.
	if m.Config.Profiles[profile].Filter.IgnoreMailErrorNotifications {
		NI := regexp.MustCompile(`An error occurred while trying to deliver the mail to the following recipients:`)
		if NI.Match(b) {
			return []byte{}, errors.New("Ignoring postal service error")
		}
	}

	// Strip out HTML header, since we don't really need it.
	// works not as intended i think
	hb := regexp.MustCompile(`(?i)<html.*/head>`)
	b = hb.ReplaceAll(b, []byte(""))

	// Try to cut out the rest of a reply that comes from Outlook.  :)
	MS := regexp.MustCompile(`<div style="border-top:solid[^>]*?><p[^>]*?><strong><span[^>]*>[A-Za-z]+:.*`)
	b = MS.ReplaceAll(b, []byte(""))

	// Eliminate Outlook iOS replies.
	OI := regexp.MustCompile(`(?s)<div class="ms-outlook-ios-signature">.*`)
	b = OI.ReplaceAll(b, []byte(""))

	// Respect to all you remaining Blackberry users.
	Bw := regexp.MustCompile(`Sent with BlackBerry Work`)
	b = Bw.ReplaceAll(b, []byte(""))

	// Try to cut out mail clients that are nice enough to tell us where the reply begins. (ProtonMail)
	om := regexp.MustCompile(`‐‐‐‐‐‐‐ Original Message ‐‐‐‐‐‐‐.*`)
	b = om.ReplaceAll(b, []byte(""))

	// Ignore namespaces (declutter html)
	xs := regexp.MustCompile(` ?xmlns:?[a-z]+?="[^"]*?"`)
	b = xs.ReplaceAll(b, []byte(""))

	// Remove ALL newlines
	nl := regexp.MustCompile(`[\r\n]+`)
	b = nl.ReplaceAll(b, []byte(""))

	// Attempt to remove replies beginning with "On X Y Z, <user@email> said:".
	ow := regexp.MustCompile(`On .*? wrote:.*`)
	b = ow.ReplaceAll(b, []byte(""))

	// Remove all <!--[MSO COMMENTS]--> and their contents. These tags come from Outlook's default mail editor.
	MC := regexp.MustCompile(`<!--\[if.*?endif]-->`)
	b = MC.ReplaceAll(b, []byte(""))

	// Remove all <!-- comments --> and their contents. Just in case there are any left.
	co := regexp.MustCompile(`<!--.*?-->`)
	b = co.ReplaceAll(b, []byte(""))

	// If someone forwards a message into the group, we'd like to hide that fact.
	fw := regexp.MustCompile(`Begin forwarded message:`)
	b = fw.ReplaceAll(b, []byte(""))

	// Remove all &nbsp;s
	nb := regexp.MustCompile(`&nbsp;?`)
	b = nb.ReplaceAll(b, []byte(""))

	// Remove all <style> tags and their contents
	st := regexp.MustCompile(`(?i)<style.*?>.*?</style>`)
	b = st.ReplaceAll(b, []byte(""))

	// Remove all <meta> tags and their contents
	me := regexp.MustCompile(`(?i)<meta.*?>.*?</meta>`)
	b = me.ReplaceAll(b, []byte(""))
	m2 := regexp.MustCompile(`(?i)<meta.*?/>`)
	b = m2.ReplaceAll(b, []byte(""))

	// Remove <div> tags and end tags (but keep the contents); Simplifying HTML
	d1 := regexp.MustCompile(`(?i)<div[^>]*?>`)
	b = d1.ReplaceAll(b, []byte(""))
	d2 := regexp.MustCompile(`(?i)</div>`)
	b = d2.ReplaceAll(b, []byte(""))

	// Remove all <o:p> tags and their contents. These are Office/Outlook-specific.
	op := regexp.MustCompile(`(?i)<o:p[^>]*>[^>]*</o:p>`)
	b = op.ReplaceAll(b, []byte(""))

	// Remove all style attributes from every tag, Markdown doesn't need these.
	sa := regexp.MustCompile(`(?i) ?style="[^"]*"`)
	b = sa.ReplaceAll(b, []byte(""))

	// Remove all style attributes from every tag, Markdown doesn't need these.
	cl := regexp.MustCompile(`(?i) ?class="[^"]*"`)
	b = cl.ReplaceAll(b, []byte(""))

	// Remove reply headers ("Subject: xxx", "From: (300 recipient list)", etc).  I've attempted to account
	// for the ways Outlook and iOS do this, but there may be other variations down the road.
	sr := regexp.MustCompile(`(?i)(<blockquote[^>]*>)?<(strong|b)>[^:]*: ?</(strong|b)> ?[^<]+<br/?>`)
	b = sr.ReplaceAll(b, []byte(""))

	// We don't care about nowrap
	nw := regexp.MustCompile(` ?nowrap="[^"]*"`)
	b = nw.ReplaceAll(b, []byte(""))

	// Remove all <span> tags and leave their contents
	sp := regexp.MustCompile(`<span[^>]*>(.*?)</span>`)
	b = sp.ReplaceAll(b, []byte("$1"))

	// Remove all <img> tags that don't point to websites.
	im := regexp.MustCompile(`<img.+src=[^h][^t][^>]*?>`)
	b = im.ReplaceAll(b, []byte(""))

	// Remove excessive <br>s
	br := regexp.MustCompile(`(<br[^>]*?>){2,}`)
	b = br.ReplaceAll(b, []byte("<br>"))

	// Simplify <td> elements, otherwise Markdown may get confused.
	tp := regexp.MustCompile(`<td([^>]*?)><p[^>]*>(.*?)</p></td>`)
	b = tp.ReplaceAll(b, []byte("<td$1>$2</td>"))

	// If we have 4 or more empty <p>s in a row, this is a good place to stop the message.  Hopefully the text below
	// that is a reply.
	MK := regexp.MustCompile(`(?i)<p></p>{4,}.*`)
	b = MK.ReplaceAll(b, []byte(""))

	// Remove any empty <p>s that remain.
	pp := regexp.MustCompile(`(?i)<p></p>`)
	b = pp.ReplaceAll(b, []byte(""))

	// Remove straggling <blockquotes>
	sb := regexp.MustCompile(`(?i)<blockquote[^>]*>$`)
	b = sb.ReplaceAll(b, []byte(""))

	// Finally, if we're lucky enough to have a "Sent from" footer to the reply, kill everything else. This is
	// typical on iOS, Samsung, and Blackberry devices. If the user has a custom signature, this won't help.
	sf := regexp.MustCompile(`Sent ([Ff]rom|via).*`)
	b = sf.ReplaceAll(b, []byte(""))

	return b, nil
}

// parseText attempts to strip everything out of the text/plain message body except for the latest reply.
//
func (m Mail2Most) parseText(b []byte) ([]byte, error) {

	// Attempt to strip the reply.
	on := regexp.MustCompile(`(?s)On .*? wrote:.*$`)
	b = on.ReplaceAll(b, []byte(""))

	// Remove forwarded-message notice.
	fw := regexp.MustCompile(`(?i)Begin forwarded message:`)
	b = fw.ReplaceAll(b, []byte(""))

	// Eliminate extra whitespace.
	ws := regexp.MustCompile(`(?s)\s{2+}`)
	b = ws.ReplaceAll(b, []byte(" "))

	// Remove reply headers.
	re := regexp.MustCompile(`(.+): ((.|\r\n\s)+)\r\n`)
	b = re.ReplaceAll(b, []byte(""))

	return b, nil
}

// parseAttachment packages an attachment in an Attachment{} type object for consumption elsewhere.
//
func (m Mail2Most) parseAttachment(body []byte, header string) (Attachment, error) {

	filename := "image"

	fn := regexp.MustCompile(`name="([^"]+)"`)
	f := fn.FindStringSubmatch(header)

	im := regexp.MustCompile(`image/([a-z]*)`)
	cn := im.FindStringSubmatch(header)
	if len(f) > 0 {
		filename = f[1]
	} else if len(cn) > 0 {
		filename = "image." + cn[1]
	}

	sum := sha256.Sum256(body)
	if _, ok := seenAttachments[sum]; ok {
		// throwing an error here is a bit strange it would be better to return the existing attachment
		// btw this could be exploited by creating a hash colision :D
		return Attachment{}, errors.New("attachment already exists")
	}

	seenAttachments[sum] = filename
	return Attachment{Filename: filename, Content: body}, nil
}
