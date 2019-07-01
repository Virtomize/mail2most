/*
Package mail2most is an implementation for reading and filtering emails and pushing them to mattermost

It uses the IMAP protocol to connect to an email account and can filter via:

- Folder
- Subject
- From
- To
- Time range

and pushes the subject and body into mattermost

*/
package mail2most
