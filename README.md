# gate-master
This is a small service which is designed to run on a Raspberry Pi within the control box of a gate mechanism (specifically designed for small gated communities). It will monitor the keypad and open the gate when the correct code is entered, but also provides a web-based interface for monitoring or configuring the gate system, allowing users to view the logs of who is entering the community, get text or email alerts, setup new PIN codes with limited access times, and more.

## Features

* Monitor the keypad at the gate, and opens the gate if a valid PIN is entered
  * Also expects to use an attached LCD to display how many characters have been entered so far. Works without it though.
  * Designed for a 4 row by 3 column keypad, with the "*" functioning as a "clear" button, and the "#" functioning as the "enter" button.
  * So you hit "1234#" to submit PIN code 1234 to open the gate, or if you mis-type a digit you can hit "*" to clear it and start over again.
* Email & Text Notifications
  * Emails are used for setting up new users and password resets
  * Whenever a PIN code is used to open the gate, an automated notification that person XYZ is entering the community will send to everybody who has registered their phone/email for those notifications.
    * These notification settings have default "Tags" for things like mail services, contractors, utility services, and more.
* Full multi-user web interface for managing gate access
  * Can login and click a button to open the gate for someone directly
* Dynamic system for creating/expiring gate PIN codes.
  * Flexible scheduling for each PIN code - only make it active certain days of the week, or between particular times of day, etc.
  * PIN codes are randomly generated, and can be 4, 6, or 8 digits long.
* Supports an attached camera at the gate, and presents that as a live video feed in the web interface so you can see who is at the gate
  * If a camera is attached, it will also snap a picture each time the gate opens and store that in the logs for review/audit later.
* Logs are recorded for each successful/failed attempt to open the gate.
  * These logs are available for viewing within the web interface (automatically prunes logs older than 1 year)
  * Additional CSV logs with JPG pictures are created within a separate directory structure (never pruned), in case you want to setup a long-term backup solution for log entries.

## Prerequisites:
These need to be installed on the Raspberry Pi before the video system will work:

* Go language meta-package
  * Run `go version` and if installed it should print out a version number.
  * Tested with Go v1.19 on a Raspberry Pi, but developed with Go v1.22+


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


# Long-form Reference Material
This section provides a much longer and more detailed walkthrough of setting up this service for non-technical users.



Fantastic guide on setting up an i2c LCD display [HERE](https://medium.com/@thedyslexiccoder/how-to-set-up-a-raspberry-pi-4-with-lcd-display-using-i2c-backpack-189a0760ae15). Include the configuration you need to do on the Raspberry Pi itself before you can configure this service and interact with the LCD screen.

## Raspberry Pi Setup Guide
This is a quick list of the steps I have used (multiple times) to setup a Raspberry Pi 4B to run this service.

1. Flash the latest 64-bit Raspberry Pi OS (Bookworm at the time I did this in 2025) to the card using the [Raspberry Pi Imager](https://www.raspberrypi.com/software/)
2. Plug the card into the Pi and do the first-boot process to finish setting up your system (username, password, localization, timezone, networking, etc)
3. Open up the Raspberry Pi Config tool in the "Preferences" menu and turn on the I2C interfaces
   * This option will be in the "Interfaces" tab.
   * It will prompt you to reboot your system to finish turning it on - skip that for now, as you will reboot in a couple minutes after setting up a few more things.
4. Open up the Add/Remove Software tool and install 2 packages:
   * Search for `golang` and install that package (version 1.19 at the time of my writing)
   * Search for `caddy` and install that package (version 2.6.2 at the time of my writing)
5. [Optional] Test that your camera is plugged in correctly and working by running `libcamera-still` from a terminal.
   * It will open a window and show the current view from the camera for a few seconds. Useful for determining if you will need to rotate the feed by 180 degrees or not.
6. Now reboot your system to finish applying the I2C config change.

### Setting up Gate-master on the Pi
Once the Pi is setup, you just need to clone and install the service, then adjust the configuration file to match your specific arrangement

#### Installing Gate-master
On the Pi, open up a terminal and run the following commands one at a time:

1. `mkdir src` (create a directory in your home folder called "src")
2. `cd src` (move into that new directory)
3. `git clone https://github.com/beanpole135/gate-master` (Download this source repository onto the Pi)
4. `cd gate-master` (move into that new directory)
5. `bash install.sh` (Build and install gate-master)
   - This step will take a while the first time you run it on a new system!! Go get some coffee or somthing and just wait until it prints out that the gate-master service is now running!

The Gate-master installation script will also setup the files to run a "Caddy" webserver instance with a sample configuration for you to finish editing. You need to install the `caddy` package first, before you can use these steps to finish it up:

1. `sudo nano /usr/local/etc/Caddyfile` to edit the configuration for Caddy
2. `sudo systemctl enable caddy` to have the system automatically start the Caddy service on system bootup
3. `sudo systemctl [start|stop|restart] caddy` to manually manage the service
4. `journalctl -u caddy` to look at the system logs for the service


#### Configuring Gate-master
Before it will really "work", you need to update the configuration file for the service based upon how you have everything setup on the Pi itself.

Top-level configuration parameters and explanations

* "host_port" : This is the local port to have the service listen on (prefixed by a colon ":"). 
  * Typically you will just point this to some random port number like ":8080", and then setup Caddy to handle your SSL certificates and reverse-proxy over to that local port based upon your domain (in case you have multiple web services running on the same system)
* "site_name" : This is just the display name that you want to show at the top of the login page for the web interface.
  * Example: "Welcome to [Your site_name here]"
* "db_file" : The local file path to where you want to place your sqlite database.
  * The default value of "/usr/local/share/gatemaster/db.sqlite" is usually fine, unless you want to store it on some other external hard drive.
* "logs_directory" : The local directory path for the persistent CSV logs with JPG images.
  * Set to a blank string to disable this functionality
* "auth" -> "jwttokensecs" : The number of seconds before forcing somebody to login again
  * Default is 3600 (1 hour) which will usually by fine for everyone

Example config:
```
    "host_url": "http://localhost:8080",
    "site_name": "MySiteName",
    "db_file": "/usr/local/share/gatemaster/db.sqlite",
    "logs_directory": "/var/log/gatemaster",
    "auth": {
        "jwttokenssecs": 3600
    },
```

Quick References:

* [RPi Pin Numbering Diagram](https://pinout.xyz/)
  * Note that the gatemaster configuration file uses GPIO pin numbers, not the "physical" pin numbers.
  * For example: physical pin 37 (at the bottom of the chart), corresponds to a GPIO pin number of 26 (so you would enter 26 into the config file)
* Run `i2cdetect -y 1` to scan I2C bus 1 and print out the hex address of your LCD device


*** Standard Steps ***
1. Run `sudo nano /usr/local/etc/gatemaster.json` within the terminal to edit the configuration file
   - There is a default config file pre-installed which you can just edit appropriately.
2. Run `sudo systemctl restart gatemaster` to restart the service after saving your changes to the config file
3. Run `journalctl -u gatemaster` to view the debugging/system logs for the service. (You will probably need to page-down to the bottom to see the most recent logs - hit "q" to quit the viewer)

*** Example configuration from my project ***
* I plugged the I2C LCD into pins 3, 4, 5, & 6 - matching the "I2C1" (I2C bus 1) labels on the pin diagram
  * Running `i2cdetect -y 1` printed out that my LCD address was "0x27" (which is a pretty common default)
```
    "lcd_i2c" : {
        "i2c_bus_number" : 1,
        "hex_address" : "0x27",
        "backlight_seconds" : 30,
    }
```

* For the camera settings, I had to flip my camera feed right-side up (camera was installed upside down), so the rotation field was set to 180. You can also set a 90 or 270 degree rotation if your camera was installed on it's side, but you will get a degraded frame rate for the video feed because the system needs to do an extra rotation of the image in post-processing.
  * The `libcamera-still` tool which tests your camera also prints a bunch of diagnostic info to the terminal so you can lookup the natural resolution of the camera as well.
```
    "camera" : {
        "rotation": 180,
        "width": 1024,
        "height": 768
    },
```

* I plugged the keypad into a bunch of open GPIO pins at the lower end of the board, and then put those GPIO pin numbers into the keypad configuration like this (note that the pins do not have to be in any particular order - I just went down the line - they just need to be accurate to the wiring on the keypad).
  * Row 1 corresponds to the "1", "2", and "3" keys
  * Row 4 corresponds to the "*", "0", and "#" keys
  * Column 1 contains the "1", "4", "7", and "*" keys
```
	4x3 Keypad
	 [1, 2, 3]
	 [4, 5, 6]
	 [7, 8, 9]
	 [*, 0, #]

Examples:
R1 + C1 = Key "1"
R2 + C3 = Key "6"
```

And how this looked in my configuration file:
```
    "keypad_pins" : {
        "row1": 7,
        "row2": 1,
        "row3": 12,
        "row4": 16,
        "col1": 8,
        "col2": 20,
        "col3": 21
    },
```



* For the gate itself, I just wired it into another open GPIO pin, and then updated the configuration file like so. The "invert drive" field can be marked "true" or "false" as needed. If your gate actuator "clicks on" each time you power up the service, then you probably need to flip the invert flag so that the default state of the actuator is in the neutral/off state.
```
    "gate" : {
        "gpio_num" : 23,
        "invert_drive" : false
    },
```
