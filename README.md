# tgbotserver

[![license](https://img.shields.io/github/license/gebv/tgbotserver.svg)]()
[![status](https://img.shields.io/badge/status-development-blue.svg)]()

Ready to use server for your telegram bot
Supported scaling

* [Quick start](#quick-start)
* [Configuration](#configuration)
* [Run](#run)
 * [mode 1 - request updates](#request-updates)
 * [mode 2 - receive updates](#receive-updates)
* [Libraries](#libraries)

# Quick start

`TGBS_TOKEN` requeired. `APP_EMAIL`  and `TGBS_WEBHOOK` optional if not use webhook.

``` bash
cat <<EOT >> env
export APP_EMAIL="myemail@mydomain.com"
export TGBS_TOKEN="123456789:ABCDEFG............................"
export TGBS_WEBHOOK="https://mydomain.com/"
export APP_LOGLEVEL="level"
EOT
source ./env
./gen_docker-compose.py
# Generation...
# Done for 'docker-compose.yml'.
# Done for 'apps/reverseproxy/config.toml'.
# OK.

docker-compose up -d pump appworker
# Creating network "tgbotserver_default" with the default driver
# Creating msgsys
# Creating database
# Creating tgbotserver_pump_1
# Creating tgbotserver_appworker_1

# if use webhook
docker-compose up -d reverseproxy appconfig appworker httplistener
```

# Configuration

* `gen_docker-compose.py` - generator settings from the `env`
* `env` - your settings
* `env.example` - template settings
* `docker-compose.yml.example` - template for docker-compose file
* `apps/reverseproxy/config.toml.example` - template for traefik settings


``` bash
cp env.example env
```

Use the file to enter your settings.
* **APP_EMAIL** - email for feedback to generate ssl ([letsencrypt.org](https://letsencrypt.org))
* **TGBS_TOKEN** - telegram bot token ([crate new bot](https://core.telegram.org/bots#create-a-new-bot))
* **TGBS_WEBHOOK** - your webhook url ([set webhook,lo](https://core.telegram.org/bots/api#setwebhook))

Update the environment variables and to generate a settings.

```
source ./env
./gen_docker-compose.py
```

# Run

Services
* **reverseproxy** - reverse proxy
* **database** - database
* **appconfig** - customize application (only if use webhook)
* **httplistener** - query processor from telegram
* **pump** - loader updates (only if not use webhook)
* **msgsys** - service messages
* **appworker** - application logic

## Request updates

If you do not have their own domain or for development on local machine should request updates.

``` bash
##################
# run server
##################

docker-compose up -d pump appworker

##################
# if required.
##################

# application scaling
# docker-compose scale appworker=5
```

## Receive updates

For production it is recommended to receive update via webhook.

``` bash
##################
# run server
##################

docker-compose up -d reverseproxy appconfig appworker httplistener

##################
# if required.
##################

# application scaling
# docker-compose scale appworker=5

# scaling listener
# docker-compose scale httplistener=10
```


# Libraries

* [docker](https://github.com/docker/docker) - container engine
* [traefik](https://github.com/containous/traefik) - reverse proxy
* [citus](https://github.com/citusdata/citus) - scalable PostgreSQL
* [nats](https://github.com/nats-io/nats) - message queues

...

# Test

``` bash
curl -XPOST -H Host:subdomain.domain.com http://127.0.0.1 -d "payload"
```

## TODO

[ ] centralized logging 