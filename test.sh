#!/bin/bash

# Directory to search for shell scripts
SEARCH_DIR=$1

# Check if directory is provided
if [ -z "$SEARCH_DIR" ]; then
  echo "Usage: $0 <directory>"
  exit 1
fi

# Find and process shell scripts containing 'pstree'
find "$SEARCH_DIR" -type f -name "*.sh" | while read -r script; do
  if grep -q 'pstree' "$script"; then
    sed -i 's/pstree/echo "pstree command replaced"/g' "$script"
    echo "Replaced 'pstree' in $script"
  fi
done