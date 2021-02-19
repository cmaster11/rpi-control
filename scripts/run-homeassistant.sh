#!/bin/bash

docker run --init -d \
  --name homeassistant \
  --restart=unless-stopped \
  -v /etc/localtime:/etc/localtime:ro \
  -v /home/pi/homeassistant:/config \
  --network=host \
  homeassistant/raspberrypi4-homeassistant:stable