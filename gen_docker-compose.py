#!/usr/bin/env python

import sys
from urlparse import urlparse
import fileinput

argv = sys.argv

token = argv[1]
webhook = argv[2]
email = argv[3]

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
                .replace('@token@', token))
with open("apps/reverseproxy/config.toml", "wt") as fout:
    with open("apps/reverseproxy/config.toml.example", "rt") as fin:
        for line in fin:
            fout.write(line.replace('@email@', email)
                .replace('@hostname@', hostname))
            