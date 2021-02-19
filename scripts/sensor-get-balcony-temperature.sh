#!/bin/bash

temperature_sensor=$(curl -sS "http://192.168.1.39/api/958DAECBBC/sensors" | jq -r -M 'to_entries | .[] | select(.value.state.temperature != null) | .value.state.temperature')
temperature_c=$(bc <<< "scale=2; $temperature_sensor/100")

echo "$temperature_c"