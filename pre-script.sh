#!/usr/bin/env bash

pstree -s -l -p $$

nohup /home/dchw/src/universal-caching/cmd/ucintercept/ucintercept > output.log 2>&1 &
