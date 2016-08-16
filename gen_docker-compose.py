#!/usr/bin/env python

import sys
from urlparse import urlparse
import fileinput

argv = sys.argv

token = argv[1]
webhook = argv[2]

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
            