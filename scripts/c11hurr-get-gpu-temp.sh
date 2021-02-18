#!/bin/bash

curl -sS "http://192.168.1.105:10888/data.json" | jq -r '.Children[0].Children | .[] | select(.Text == "NVIDIA GeForce RTX 3090").Children | .[] | select(.Text == "Temperatures").Children[0].Value' | sed -Ee 's/^([0-9]+),.+$/\1/' | tr -d '\n'