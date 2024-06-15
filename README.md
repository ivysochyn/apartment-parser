# ApartiBot - Your personal apartment search assistant
[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit)](https://github.com/pre-commit/pre-commit)

ApartiBot is a Telegram bot that aggregates apartment listings from different sources and sends them to you.
The bot is currently in development and only supports the following sources:
* [OLX](https://www.olx.pl/)
* [Otodom](https://www.otodom.pl/)

## How to use

To start using the bot, you need to have a Telegram Bot Token.
The token is fetched from the environment variable `TELEGRAM_APITOKEN`:
```bash
TELEGRAM_APITOKEN=your_token_here go run apartment-parser
```

## Systemd

In order to run the bot as a systemd service, you need to create a service file in the `/etc/systemd/system/` directory with `<name>.service` name:
```ini
[Unit]
Description=Binary to start a telegram bot responsible for parsing Olx and Otodom apartment offers.

Wants=network.target
After=syslog.target network-online.target

[Service]
Type=simple
Environment="TELEGRAM_APITOKEN=<TOKEN>"
ExecStart=<BINARY_PATH>
Restart=on-failure
RestartSec=10
KillMode=process

[Install]
WantedBy=multi-user.target
```

Make sure to replace `<TOKEN>` with your Telegram Bot Token and `<BINARY_PATH>` with the path to the binary file.

You can verify if `systemd` can find the service by listing all available services:
```bash
sudo systemctl list-units --type=service
```

Also, you can check the status of the service:
```bash
sudo systemctl status <name>.service
```

After creating the service file, you need to reload the systemd daemon and start the service:
```bash
sudo systemctl daemon-reload
sudo systemctl start <name>.service
```

Also, you can enable the service to start on boot:
```bash
sudo systemctl enable <name>.service
```
