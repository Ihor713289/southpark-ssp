import json
import os
import signal
import sys
import time

import pika
from dotenv import load_dotenv
from pika.exceptions import AMQPConnectionError

load_dotenv()

QUEUE_NAME = "southpark_messages"
RABBITMQ_URL = os.getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")


def connect(attempts=10, delay=3):
    for attempt in range(1, attempts + 1):
        try:
            return pika.BlockingConnection(pika.URLParameters(RABBITMQ_URL))
        except AMQPConnectionError as err:
            print(f"RabbitMQ not ready (attempt {attempt}/{attempts}): {err!r}", flush=True)
            time.sleep(delay)
    sys.exit("Could not connect to RabbitMQ")


def on_message(channel, method, properties, body):
    try:
        message = json.loads(body)
        print(f"{message.get('author', 'Unknown')}: {message.get('body', '')}", flush=True)
    except (json.JSONDecodeError, AttributeError):
        print(f"Received malformed message: {body!r}", flush=True)
    channel.basic_ack(delivery_tag=method.delivery_tag)


def main():
    connection = connect()
    channel = connection.channel()
    # Must match the declaration in the Go publisher (durable, non-exclusive, no auto-delete).
    channel.queue_declare(queue=QUEUE_NAME, durable=True)
    channel.basic_qos(prefetch_count=1)
    channel.basic_consume(queue=QUEUE_NAME, on_message_callback=on_message)

    signal.signal(signal.SIGTERM, lambda *_: channel.stop_consuming())

    print(f"Waiting for messages on '{QUEUE_NAME}'...", flush=True)
    try:
        channel.start_consuming()
    except KeyboardInterrupt:
        channel.stop_consuming()
    finally:
        connection.close()


if __name__ == "__main__":
    main()
