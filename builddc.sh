#!/bin/bash

cp docker-compose.yml.example docker-compose.yml

DOMAIN=$1
ENTRYPOINT=$2
DOMAIN=${DOMAIN:-example.com}
ENTRYPOINT=${ENTRYPOINT:-entrypoint}

sed -i -- 's/@domain@/'$DOMAIN'/g' docker-compose.yml
sed -i -- 's/@entrypoint@/'$ENTRYPOINT'/g' docker-compose.yml