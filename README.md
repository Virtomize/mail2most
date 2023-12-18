[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=VBXHBYFU44T5W&source=url)
[![GoDoc](https://img.shields.io/badge/godoc-reference-green.svg)](https://godoc.org/github.com/virtomize/mail2most/lib)
[![Go Report Card](https://goreportcard.com/badge/github.com/virtomize/mail2most)](https://goreportcard.com/report/github.com/virtomize/mail2most)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/virtomize/mail2most/blob/master/LICENSE)
[![Built with Mage](https://magefile.org/badge.svg)](https://magefile.org)
[![Build Status](https://api.travis-ci.com/virtomize/mail2most.svg?branch=master)](https://travis-ci.com/virtomize/mail2most)

# [![Mail2Most](https://user-images.githubusercontent.com/13348918/60418882-560c3480-9be4-11e9-9f30-b0124a162630.png)](https://github.com/virtomize/mail2most)

Filter emails from mail accounts and send them to mattermost.

![mail2most-image](https://user-images.githubusercontent.com/13348918/60437141-ff1b5500-9c0d-11e9-913f-ae7c4a034b10.png)

# Features

- IMAP(S) support
- StartTLS support
- Mattermost v4 API support
- HTML 2 Markdown support
- Filter mails by Folder
- Filter mails by From
- Filter mails by To
- Filter mails by Subject
- Filter mails by TimeRange
- Mattermost broadcasts
- Choose to post Subject and Body or Subject only
- Send to channels and/or users
- Profile management including default profiles
- Mail attachment support

Missing feature or found a bug ? Feel free to open an [issue](https://github.com/virtomize/mail2most/issues) and let us know !

## Donation

If this project helps you, feel free to give us a cup of coffee :).

[![paypal](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=VBXHBYFU44T5W&source=url)

# Install

## download

Download [Latest Release Version](https://github.com/virtomize/mail2most/releases/latest)

## build it yourself

You can compile the project yourself using this repo and [mage](https://magefile.org).
Just clone the repo and run `mage build`, you can find the binary under `bin/mail2most`

# Usage

- create a mattermost user
- create or use an existsing email user to connect to your mail server via IMAP
- edit `conf/mail2most.conf` and configure your mail and mattermost credentials
- configure your filters
- run Mail2Most `./mail2most` or with config path `./mail2most -c conf/mail2most.conf`

# Version Info

- run `./mail2most -version` to get current version information

## example conf - filter descriptions

**just configure the filters you need if a filter is not defined it is not used !**

see [example configuration](https://github.com/virtomize/mail2most/blob/master/conf/mail2most.conf) for more details.

# Run Mail2Most as a service

You can run Mail2Most using docker, docker-compose or as a systemd service.

## docker

[docker-hub link](https://hub.docker.com/r/virtomize/mail2most)

Using docker you need to change the path to your mail2most.conf

```
docker run \
  -v /path/to/mail2most.conf:/mail2most/conf/mail2most.conf \
  virtomize/mail2most:latest
```

e.g. if you are in this repo:

```
docker run \
  -v $(pwd)/conf/mail2most.conf:/mail2most/conf/mail2most.conf \
  virtomize/mail2most:latest
```

## docker-compose

[docker-hub link](https://hub.docker.com/r/virtomize/mail2most)

Using docker-compose you can just edit the `conf/mail2most.conf` or change the path inside the docker-compose.yml to your config:

```
    volumes:
      - ./conf/mail2most.conf:/mail2most/conf/mail2most.conf
```

needs to be changed to

```
    volumes:
      - /path/to/my/mail2most.conf:/mail2most/conf/mail2most.conf
```

then just start a container user

```
docker-compose up -d
```

## Systemd

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
RestartSec=60

[Install]
WantedBy=multi-user.target
```

enable and start using

```
systemctl enable mail2most
systemctl start mail2most
```

## Common problems

**Problem: mail2most crashes after profile changes**

Solution: This happens when the data.json is not consistent to the config changes. Delete data.json to solve this problem.

**Problem: Channel contains special characters mattermost can not found the channel**

Solution: Mattermost does not support special characters for channel names, only in display names. To find the correct channel name use the last part of the url found under `view info`

# Contribution to Mail2Most

Thank you for participating to this project.
Please see our [Contribution Guidlines](https://github.com/virtomize/mail2most/blob/master/CONTRIBUTING.md) for more information.
