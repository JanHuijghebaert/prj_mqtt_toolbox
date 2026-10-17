package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
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
	clientID := flag.String("client-id", "", "MQTT client ID (generated automatically when omitted)")
	username := flag.String("username", "", "MQTT username")
	password := flag.String("password", "", "MQTT password")
	topic := flag.String("topic", "test/topic", "topic to subscribe to or publish on")
	qos := flag.Uint("qos", 1, "message QoS (0, 1, or 2)")
	publish := flag.String("publish", "", "publish this message to the topic and exit")
	flag.Parse()

	if *qos > 2 {
		return fmt.Errorf("invalid QoS: must be 0, 1, or 2")
	}

	if *clientID == "" {
		randomNumber, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
		if err != nil {
			return fmt.Errorf("failed to generate MQTT client ID: %w", err)
		}
		*clientID = fmt.Sprintf("clId%06d", randomNumber.Int64())
	}
	fmt.Fprintf(os.Stderr, "Using MQTT client ID: %s\n", *clientID)

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
	fmt.Fprintf(os.Stderr, "Connecting to MQTT broker %s...\n", *broker)
	token := client.Connect()
	token.Wait()
	if err := token.Error(); err != nil {
		return fmt.Errorf("MQTT connection failed: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Connected to MQTT broker %s\n", *broker)
	defer func() {
		fmt.Fprintln(os.Stderr, "Disconnecting from MQTT broker...")
		client.Disconnect(250)
	}()

	if *publish != "" {
		fmt.Fprintf(os.Stderr, "Publishing to %s (QoS %d)...\n", *topic, *qos)
		token := client.Publish(*topic, byte(*qos), false, *publish)
		token.Wait()
		if err := token.Error(); err != nil {
			return fmt.Errorf("MQTT publish failed: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Published to %s\n", *topic)
		return nil
	}

	fmt.Fprintf(os.Stderr, "Subscribing to %s (QoS %d)...\n", *topic, *qos)
	token = client.Subscribe(*topic, byte(*qos), nil)
	token.Wait()
	if err := token.Error(); err != nil {
		return fmt.Errorf("MQTT subscription failed: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Subscription to %s confirmed\n", *topic)
	fmt.Printf("Subscribed to %s on %s; press Ctrl+C to exit\n", *topic, *broker)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)
	<-interrupt
	return nil
}
