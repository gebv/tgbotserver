# tgbotserver

[![license](https://img.shields.io/github/license/gebv/tgbotserver.svg)]()
[![status](https://img.shields.io/badge/status-development-blue.svg)]()

Ready to use server for your telegram bot
Supported scaling

* [Configuration](#configuration)
* [Run](#run)
* [Overview](#overview)

# Configuration

For generate `docker-compose.yml` file use the script, follow command:

```
./gen_docker-compose.py 1234:ABCD "https://sub1.sub2.domain.com/path1/path2/webhook"
```

* First argument it is token of telegram bot. 
* Second argument it is your webhook url. 

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