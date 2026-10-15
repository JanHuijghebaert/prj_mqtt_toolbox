package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	broker := flag.String("broker", "tcp://localhost:1883", "MQTT broker URL")
	clientID := flag.String("client-id", "mqtt-toolbox-client", "MQTT client ID")
	username := flag.String("username", "", "MQTT username")
	password := flag.String("password", "", "MQTT password")
	topic := flag.String("topic", "test/topic", "topic to subscribe to or publish on")
	qos := flag.Uint("qos", 1, "message QoS (0, 1, or 2)")
	publish := flag.String("publish", "", "publish this message to the topic and exit")
	flag.Parse()

	if *qos > 2 {
		return fmt.Errorf("invalid QoS: must be 0, 1, or 2")
	}

	options := MQTT.NewClientOptions().
		AddBroker(*broker).
		SetClientID(*clientID).
		SetUsername(*username).
		SetPassword(*password).
		SetCleanSession(true).
		SetConnectTimeout(10 * time.Second).
		SetConnectionLostHandler(func(_ MQTT.Client, err error) {
			fmt.Fprintf(os.Stderr, "MQTT connection lost: %v\n", err)
		})

	if *publish == "" {
		options.SetDefaultPublishHandler(func(_ MQTT.Client, message MQTT.Message) {
			fmt.Printf("%s: %s\n", message.Topic(), message.Payload())
		})
	}

	client := MQTT.NewClient(options)
	token := client.Connect()
	token.Wait()
	if err := token.Error(); err != nil {
		return fmt.Errorf("MQTT connection failed: %w", err)
	}
	defer client.Disconnect(250)

	if *publish != "" {
		token := client.Publish(*topic, byte(*qos), false, *publish)
		token.Wait()
		if err := token.Error(); err != nil {
			return fmt.Errorf("MQTT publish failed: %w", err)
		}
		return nil
	}

	token = client.Subscribe(*topic, byte(*qos), nil)
	token.Wait()
	if err := token.Error(); err != nil {
		return fmt.Errorf("MQTT subscription failed: %w", err)
	}
	fmt.Printf("Subscribed to %s on %s; press Ctrl+C to exit\n", *topic, *broker)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)
	<-interrupt
	return nil
}
