#!/usr/bin/env bash

sudo nohup ./interceptor &

echo "idk just execute shit"

while true
do
  python3 -c 'import time; time.sleep(2)'
  sleep 1
  echo "I did it"
done
