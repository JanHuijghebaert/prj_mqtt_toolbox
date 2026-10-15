# Mosquitto

## Setup

### Install

```sh
$ sudo apt update
$ sudo apt install mosquitto mosquitto-clients -y
```

### Start service

```sh
$ sudo systemctl enable mosquitto
$ sudo systemctl start mosquitto
```

### Configure

For network & security, open the configuration file ```sudo nano /etc/mosquitto/mosquitto.conf```. Add the following lines to the bottom of the file.

```sh
...
# Listen to port 1883 on local network
listener 1883 0.0.0.0

# Only allow connections with credentials
allow_anonymous false

# Path to password file
password_file /etc/mosquitto/passwd
```

Make sure the mosquitto service has ownership of the password file.

```sh
$ sudo chown mosquitto:mosquitto /etc/mosquitto/passwd
$ sudo chmod 0600 /etc/mosquitto/passwd
```

Create login credentials.

```sh
$ sudo mosquitto_passwd -c /etc/mosquitto/passwd <your_username>
Password: <your_password>
Reenter password: <your_password>
```

### Restart service

```sh
$ sudo systemctl restart mosquitto.service
$ sudo systemctl status mosquitto.service
```

## Test

### On localhost

Terminal 1 - Subscribe to topic:

```sh
$ mosquitto_sub -h localhost -u <your_username> -P <your_password> -t "topic/test"
```

Terminal 2 - Publish message to topic:

```sh
$ mosquitto_pub -h localhost -u <your_username> -P <your_password> -t "topic/test" -m "Payload message"
```

### On other machine in local network

On MQTT Explorer:

- Protocol: mqtt://
- Host: <ip_addres_of_broker>
- Username: <your_username>
- Password: <your_password>
