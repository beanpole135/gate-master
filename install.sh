#!/bin/bash

# Simple installation script for putting Gatemaster onto the Raspberry Pi
# This will do either a clean install or an update of an existing install
# NOTE: This will run sudo occasionally to do the installations

# Do not edit these without also editing the systemd service file!
binpath="/usr/bin/"
instpath="/usr/local/bin/"
confpath="/usr/local/etc/"
servicedir="/etc/systemd/system/"

# Check for package dependencies
ok=1
for exe in go v4l2-ctl
do
  if [[ ! -f ${binpath}${exe} || ! -x ${binpath}${exe} ]] ; then
    echo "Missing Utility (${exe}): Please install it first"
    ok=2
  fi
done
if [ ${ok} -eq 2 ] ; then
  exit 1
fi

# Build the executable
(cd src-go && go get && go build -o gatemaster)
if [ $? -ne 0 ] ; then
  echo "Error building gatemaster executable"
  exit 1
fi


# Install the executable
sudo systemctl stop gatemaster #in case it is running right now
sudo cp src-go/gatemaster ${instpath}gatemaster
if [ $? -ne 0 ] ; then
    echo "Error installing gatemaster executable"
    exit 1
fi
# Install the config file (if not already exists)
if [ ! -f "${confpath}gatemaster.json" ] ; then
    sudo cp src-go/config.json.sample "${confpath}gatemaster.json"
fi
if [ ! -f "${confpath}Caddyfile" ] ; then
    sudo cp systemd/Caddyfile.sample "${confpath}Caddyfile"
fi
# Always replace the sample config file
sudo cp src-go/config.json.sample "${confpath}gatemaster.json.sample"

# Install the service file
sudo cp systemd/gatemaster.service "${servicedir}gatemaster.service"
sudo cp systemd/caddy.service "${servicedir}caddy.service"

# Enable/restart the service
sudo systemctl enable gatemaster
sudo systemctl restart gatemaster

echo "gatemaster successfully installed and should be running now!"

