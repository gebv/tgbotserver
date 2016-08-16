# tgbotserver

[![license](https://img.shields.io/github/license/gebv/tgbotserver.svg?maxAge=2592000)]()

Ready to use server for your telegram bot
Supported scaling

* [Configuration](#configuration)
* [Run](#run)
* [Overview](#overview)

# Configuration

Build `docker-compose.yml` file.
For build `docker-compose.yml` file use the script, follow command:
```
./builddc.sh domain_name subdomain_name
```

First argument it is host name of your server.
Second argument it is subdomain name - entrypoint for webhook for telegram bot.

# Run

``` bash
# run server
docker-compose up -d

# application scaling
docker-compose scale appworker=5

# scaling listener
docker-compose scale httplistener=10

# scaling database
docker-compose scale dbworker=3
```

# Overview

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

[x] scaling
[ ] centralized logging 