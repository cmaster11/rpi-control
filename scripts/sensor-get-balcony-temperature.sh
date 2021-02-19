#!/bin/bash

temperature_k=$(curl -sS "http://192.168.1.39/api/958DAECBBC/sensors" | jq -r -M 'to_entries | .[] | select(.value.state.temperature != null) | .value.state.temperature')
temperature_c=$((temperature_k + 273))

echo "$temperature_c"