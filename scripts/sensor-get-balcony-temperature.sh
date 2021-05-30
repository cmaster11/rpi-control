#!/bin/bash

temperature_sensor=$(curl -sS "http://192.168.1.39/api/864A633FD5/sensors" | jq -r -M 'to_entries | .[] | select(.value.state.temperature != null) | .value.state.temperature')
temperature_c=$(bc <<< "scale=1; $temperature_sensor/100")

echo "$temperature_c"