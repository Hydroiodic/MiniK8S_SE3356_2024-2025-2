#!/bin/bash

# Get the directory of the current script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Prepare folders for DNS example
mkdir -p /tmp/dns/web
cp -r $DIR/nginx /tmp/dns/
cp -r $DIR/web /tmp/dns/web/
