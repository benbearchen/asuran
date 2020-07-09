#!/bin/sh

if [ ! -e "server.key" ]; then
    echo "create new server.key..."
    openssl genrsa -out server.key 2048
else
    echo "server.key already exist"
fi

if [ ! -e "server.cert" ]; then
    echo "create new server.cert..."
    openssl req -new -x509 -key server.key -out server.cert -days 3650
else
    echo "server.cert already exist"
fi

