#!/usr/bin/env python

import os
from urlparse import urlparse
import fileinput

token = os.environ['TGBS_TOKEN']
webhook = os.environ['TGBS_WEBHOOK']
email = os.environ['APP_EMAIL']
loglevel = os.environ['APP_LOGLEVEL']

print("Generation...")

result = urlparse(webhook)

hostname = result.hostname
domainArr = hostname.split('.')

# only for top-level domains
domain = '.'.join(domainArr[len(domainArr)-2:len(domainArr)]) 
path = result.path

with open("docker-compose.yml", "wt") as fout:
    with open("docker-compose.yml.example", "rt") as fin:
        for line in fin:
            fout.write(line.replace('@domain@', domain)
                .replace('@webhook@', webhook)
                .replace('@hostname@', hostname)
                .replace('@token@', token)
                .replace('@appLogLevel@', loglevel))
print("Done for 'docker-compose.yml'.")

with open("apps/reverseproxy/config.toml", "wt") as fout:
    with open("apps/reverseproxy/config.toml.example", "rt") as fin:
        for line in fin:
            fout.write(line.replace('@email@', email)
                .replace('@hostname@', hostname))
print("Done for 'apps/reverseproxy/config.toml'.")
print("OK.")
            