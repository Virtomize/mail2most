[![GoDoc](https://img.shields.io/badge/godoc-reference-green.svg)](https://godoc.org/github.com/cseeger-epages/mail2most/lib)
[![Go Report Card](https://goreportcard.com/badge/github.com/cseeger-epages/mail2most)](https://goreportcard.com/report/github.com/cseeger-epages/mail2most)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/cseeger-epages/mail2most/blob/master/LICENSE)
[![Built with Mage](https://magefile.org/badge.svg)](https://magefile.org)
[![Build Status](https://api.travis-ci.org/cseeger-epages/mail2most.svg?branch=master)](https://travis-ci.org/cseeger-epages/mail2most)

# [![Mail2Most](https://user-images.githubusercontent.com/13348918/60418882-560c3480-9be4-11e9-9f30-b0124a162630.png)](https://github.com/cseeger-epages/mail2most)

Filter emails from mail accounts and send them to mattermost.

![mail2most-image](https://user-images.githubusercontent.com/13348918/60437141-ff1b5500-9c0d-11e9-913f-ae7c4a034b10.png)

# Features

- IMAP support
- Mattermost v4 API support
- Filter mails by Folder
- Filter mails by From
- Filter mails by To
- Filter mails by Subject
- Filter mails by TimeRange
- Choose to post Subject and Body or Subject only

Missing feature or found a bug ? Feel free to open an [issue](https://github.com/cseeger-epages/mail2most/issues) and let us know !

# Install

## build it yourself

You can compile the project yourself using this repo and [mage](https://magefile.org).
Just clone the repo and run `mage build`, you can find the binary under `bin/mail2most`

## download

Download [Latest Release Version](https://github.com/cseeger-epages/mail2most/releases/latest)

# Usage

- create a mattermost user 
- create or use an existsing email user to connect to your mail server via IMAP
- edit `conf/mail2most.conf` and configure your mail and mattermost credentials
- configure your filters
- run Mail2Most `./mail2most` or with config path `./mail2most -c conf/mail2most.conf`

## example conf filter description

see [example configuration](https://github.com/cseeger-epages/mail2most/blob/master/conf/mail2most.conf)

# Systemd example configuration

Systemd allows you to create a background service to run mail2most managed by your system:

- create `/opt/mail2most` and place the mail2most binary into it
  - `mkdir -p /opt/mail2most/conf`
- create `/opt/mail2most/conf/mail2most.conf`
- place the following file to `/etc/systemd/system/mail2most.service`

```
# mail2most
[Unit]
Description=mail2most

[Service]
Type=simple
WorkingDirectory=/opt/mail2most
ExecStart=/opt/mail2most/mail2most -c conf/mail2most.conf
Restart=always
Nice=5

[Install]
WantedBy=multi-user.target
```

enable and start using

```
systemctl enable mail2most
systemctl start mail2most
```

# Contribution to Mail2Most

Thank you for participating to this project.
Please see our [Contribution Guidlines](https://github.com/cseeger-epages/mail2most/blob/master/CONTRIBUTING.md) for more information.
