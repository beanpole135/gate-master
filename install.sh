#!/bin/bash

# Simple installation script for putting Gatemaster onto the Raspberry Pi
# This will do either a clean install or an update of an existing install
# NOTE: This will run sudo occasionally to do the installations

# Do not edit these without also editing the systemd service file!
binpath="/usr/bin/"
instpath="/usr/local/bin/"
confpath="/usr/local/etc/"
servicefile="/etc/systemd/system/gatemaster.service"

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
cd src-go && go get && go build -o gatemaster
if [ $? -ne 0 ] ; then
  echo "Error building gatemaster executable"
  exit 1
fi

function run_as_root() {
# Install the executable
    cp src-go/gatemaster ${instpath}gatemaster
    if [ $? -ne 0 ] ; then
        echo "Error installing gatemaster executable"
        exit 1
    fi
    # Install the config file (if not already exists)
    if [ ! -f "${confpath}gatemaster.json" ] ; then
        cp src-go/config.json.sample "${confpath}gatemaster.json"
    fi
    # Always replace the sample config file
    cp src-go/config.json.sample "${confpath}gatemaster.json.sample"

    # Install the service file
    cp systemd/gatemaster.service "${servicefile}"

    # Enable/restart the service
    systemctl enable gatemaster
    systemctl restart gatemaster
}

# Now run all the operations that need to run with root permissions to install
sudo bash -c "$(declare -f run_as_root); run_as_root"
