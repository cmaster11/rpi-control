#!/bin/bash

if nc -zv 192.168.1.109 22 2>/dev/null; then
    echo "Yes!"
else
    echo "No!"
fi