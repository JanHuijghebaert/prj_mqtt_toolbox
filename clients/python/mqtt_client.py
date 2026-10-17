#!/usr/bin/env python3

import argparse
import sys
import threading
from urllib.parse import urlsplit

import paho.mqtt.client as mqtt

CONNECTION_TIMEOUT = 10

def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Subscribe to or publish an MQTT topic")
    parser.add_argument(
        "--broker", default="tcp://localhost:1883", help="MQTT broker URL"
    )
    parser.add_argument(
        "--client-id", default="mqtt-toolbox-client", help="MQTT client ID"
    )
    parser.add_argument("--username", default="", help="MQTT username")
    parser.add_argument("--password", default="", help="MQTT password")
    parser.add_argument(
        "--topic", default="test/topic", help="topic to subscribe to or publish on"
    )
    parser.add_argument(
        "--qos", type=int, choices=(0, 1, 2), default=1, help="message QoS"
    )
    parser.add_argument(
        "--publish", default="", help="publish this message to the topic and exit"
    )
    return parser.parse_args()


def broker_connection(broker_url: str) -> tuple[str, int, bool]:
    try:
        broker = urlsplit(broker_url)
        port = broker.port
    except ValueError as error:
        raise ValueError(f"invalid MQTT broker URL: {error}") from error

    if broker.scheme not in ("tcp", "mqtt", "ssl", "mqtts"):
        raise ValueError("broker URL scheme must be tcp, mqtt, ssl, or mqtts")
    if not broker.hostname or broker.username or broker.password:
        raise ValueError("broker URL must contain a host and no credentials")
    if broker.path not in ("", "/") or broker.query or broker.fragment:
        raise ValueError("broker URL must not contain a path, query, or fragment")

    use_tls = broker.scheme in ("ssl", "mqtts")
    return broker.hostname, port or (8883 if use_tls else 1883), use_tls


def run(args: argparse.Namespace) -> None:
    host, port, use_tls = broker_connection(args.broker)
    client = mqtt.Client(
        callback_api_version=mqtt.CallbackAPIVersion.VERSION2,
        client_id=args.client_id,
        clean_session=True,
    )
    client.socket_timeout = CONNECTION_TIMEOUT
    if args.username or args.password:
        client.username_pw_set(args.username or None, args.password or None)
    if use_tls:
        client.tls_set()

    connected = threading.Event()
    subscribed = threading.Event()
    connection_error: list[str] = []
    subscription_error: list[str] = []

    def on_connect(
        _client: mqtt.Client,
        _userdata: object,
        _flags: mqtt.ConnectFlags,
        reason_code: mqtt.ReasonCode,
        _properties: mqtt.Properties | None,
    ) -> None:
        if reason_code.is_failure:
            connection_error.append(str(reason_code))
        connected.set()

    def on_disconnect(
        _client: mqtt.Client,
        _userdata: object,
        _disconnect_flags: mqtt.DisconnectFlags,
        reason_code: mqtt.ReasonCode,
        _properties: mqtt.Properties | None,
    ) -> None:
        if reason_code.is_failure:
            print(f"MQTT connection lost: {reason_code}", file=sys.stderr)

    def on_message(
        _client: mqtt.Client, _userdata: object, message: mqtt.MQTTMessage
    ) -> None:
        payload = message.payload.decode("utf-8", errors="replace")
        print(f"{message.topic}: {payload}", flush=True)

    def on_subscribe(
        _client: mqtt.Client,
        _userdata: object,
        _mid: int,
        reason_codes: list[mqtt.ReasonCode],
        _properties: mqtt.Properties | None,
    ) -> None:
        if any(reason_code.is_failure for reason_code in reason_codes):
            subscription_error.append(", ".join(map(str, reason_codes)))
        subscribed.set()

    client.on_connect = on_connect
    client.on_disconnect = on_disconnect
    client.on_message = on_message
    client.on_subscribe = on_subscribe

    try:
        client.connect(host, port, keepalive=60)
        client.loop_start()
        if not connected.wait(CONNECTION_TIMEOUT):
            raise RuntimeError("MQTT connection timed out")
        if connection_error:
            raise RuntimeError(f"MQTT connection failed: {connection_error[0]}")

        if args.publish:
            info = client.publish(args.topic, args.publish, qos=args.qos)
            if info.rc != mqtt.MQTT_ERR_SUCCESS:
                raise RuntimeError(f"MQTT publish failed: {mqtt.error_string(info.rc)}")
            info.wait_for_publish(timeout=CONNECTION_TIMEOUT)
            if not info.is_published():
                raise RuntimeError("MQTT publish timed out")
            return

        result, _mid = client.subscribe(args.topic, qos=args.qos)
        if result != mqtt.MQTT_ERR_SUCCESS:
            raise RuntimeError(f"MQTT subscription failed: {mqtt.error_string(result)}")
        if not subscribed.wait(CONNECTION_TIMEOUT):
            raise RuntimeError("MQTT subscription timed out")
        if subscription_error:
            raise RuntimeError(f"MQTT subscription failed: {subscription_error[0]}")

        print(
            f"Subscribed to {args.topic} on {args.broker}; press Ctrl+C to exit",
            flush=True,
        )
        threading.Event().wait()
    finally:
        if client.is_connected():
            client.disconnect()
        client.loop_stop()


def main() -> int:
    args = parse_args()
    try:
        run(args)
    except (OSError, RuntimeError, ValueError) as error:
        print(error, file=sys.stderr)
        return 1
    except KeyboardInterrupt:
        pass
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
