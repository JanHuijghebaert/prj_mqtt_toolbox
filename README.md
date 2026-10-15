# MQTT toolbox

## Brokers

- Mosquitto (tested on RPi)

## Clients

As free MQTT client, I mostly use [MQTT Explorer](https://mqtt-explorer.com/). But for scripting:

- Python client (todo)
- Golang client

### Golang client

Run the client from `clients/golang`. By default it subscribes to `test/topic`
on `tcp://localhost:1883` and prints received messages:

```sh
go run . -broker tcp://localhost:1883 -topic test/topic
```

Publish a message instead, then exit:

```sh
go run . -broker tcp://localhost:1883 -topic test/topic -publish "hello MQTT"
```

Use `-username`, `-password`, and `-qos` to configure broker credentials and
message QoS. QoS must be 0, 1, or 2.
