# gate-master
This is a small service which is designed to run on a Raspberry Pi within the control box of a gate mechanism (specifically designed for small gated communities). It will monitor the keypad and open the gate when the correct code is entered, but also provides a web-based interface for monitoring or configuring the gate system, allowing users to view the logs of who is entering the community, get text or email alerts, setup new PIN codes with limited access times, and more.


## Prerequisites:
These need to be installed on the Raspberry Pi before the video system will work:

* Go language meta-package
  * Run `go version` and if installed it should print out a version number.
  * Tested with Go v1.19 on a Raspberry Pi, but developed with Go v1.22+
* `sudo apt-get install v4l-utils`
  * Run `v4l2-ctl --list-devices` to ensure the "/dev/video0" device name shows up now.


## Building for the Raspberry Pi
Clone the repository onto the Pi itself, and then move into that directory and run the following commands from the terminal:

* Run `bash install.sh`
  * Note that this will use `sudo` to complete the installation and places files into:
    * /usr/local/bin/gatemaster
    * /usr/local/etc/gatemaster.json (Primary config file)
    * /usr/local/etc/gatemaster.json.sample (Sample of config file template if you need it)
* Edit the primary configuration file and populate all the settings with your specific configuration values.
* Run `sudo systemctl restart gatemaster` to restart the service after you are finished changing the config file (required to pick up on the config changes)
* Within a browser, open up http://localhost:8080 (or alternate port if that is what you configured) in order to launch the application interface.
* Once you are happy with how it is configured, you will probably want to install and turn on a webserver to allow access to the localhost port. (I recommend a quick [Caddy](https://caddyserver.com/) instance)

***LOGIN NOTE*** : For a new database (no defined users), simply login with "admin" as the email & anything as the password, and it will let you login. Create your first user account with admin permissions and that default admin access will be disabled from then on (use the new user you created to login after that).

## Upgrading the version
* Git pull the source repo
* Re-run the `install.sh` script. Will not overwrite your configuration file or database!
* Compare your current configuration file to the sample one in `/usr/local/etc`. If there are new fields in the sample file, you may want to set those up in your primary config file and restart the service again.


## Reference Material

Fantastic guide on setting up an i2c LCD display [HERE](https://medium.com/@thedyslexiccoder/how-to-set-up-a-raspberry-pi-4-with-lcd-display-using-i2c-backpack-189a0760ae15). Include the configuration you need to do on the Raspberry Pi itself before you can configure this service and interact with the LCD screen.

## Requirements

