#!/bin/bash

curl -sS "http://192.168.1.105:10888/data.json" | jq -r '.Children[0].Children | .[] | select(.Text == "AMD Ryzen 7 3700X").Children | .[] | select(.Text == "Temperatures").Children[0].Value' | sed -Ee 's/^([0-9]+),.+$/\1/'