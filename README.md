# gate-master
Gate control systems


## Prerequisites:
These need to be installed on the Raspberry Pi before the video system will work:

* `sudo apt-get install v4l-utils`
  * Run `v4l2-ctl --list-devices` to ensure the "/dev/video0" device name shows up now.


## Building for the Raspberry Pi
`GOOS=linux GOARCH=arm GOARM=7 go build`