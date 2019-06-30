[![Built with Mage](https://magefile.org/badge.svg)](https://magefile.org)

# Mail2Most

Filter emails from mail accounts and send them to mattermost.

Features:
- Imap
- Mattermost v4
- Filter by Folder
- Filter by From
- Filter by To
- Filter by Subject
- Filter by TimeRange

Todo:
- tests

# example conf

see [example configuration](https://github.com/cseeger-epages/mail2most/blob/master/conf/mail2most.conf)

# Systemd example configuration

create `/opt/mail2most` and place the mail2most binary into it.
create `/opt/mail2most/conf/mail2most.conf`

place the following file to `/etc/systemd/system/mail2most.service`

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
