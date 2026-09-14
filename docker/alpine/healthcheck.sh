#!/bin/sh

set -e

PORT=${FB_PORT:-$(sh /JSON.sh </config/settings.json | grep '\["port"\]' | awk '{print $2}')}
ADDRESS=${FB_ADDRESS:-$(sh /JSON.sh </config/settings.json | grep '\["address"\]' | awk '{print $2}' | sed 's/"//g')}
ADDRESS=${ADDRESS:-localhost}

wget -q --spider "http://$ADDRESS:$PORT/health" || exit 1
