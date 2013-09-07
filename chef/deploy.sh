#!/bin/bash

# Usage: ./deploy.sh [host]

# check against dirty working tree. maybe eventually use git rather than rsync,
# but that will make bootstrapping more difficult
rsync -avz . $1:~/chef --exclude .git
echo "Running install.sh..."
ssh -t "$1" 'cd ~/chef && sudo bash install.sh'

