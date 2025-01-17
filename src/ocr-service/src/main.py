import logging
import pika
import time
import json
import ocr_processing

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)

def process_message(ch, method, properties, body, ack_channel):
    logging.info("Callback triggered. Received message...")
    try:
        message = json.loads(body)
        logging.info(f"Message received: {message}")
        file_path = message.get("file_path")
        meeting_id = message.get("meeting_id")
        task_id = properties.correlation_id

        if not file_path or not meeting_id:
            logging.warning("Invalid message format. Skipping...")
            if ch.is_open:
                ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
            else:
                logging.warning("Channel is closed. Cannot nack message.")
            return

        logging.info(f"Received task for file: {file_path}, meeting ID: {meeting_id}")

        ocr_processing.ocr_image(file_path=file_path, meeting_id=meeting_id)

        ack_message = {
            "meeting_id": meeting_id,
            "task_id": task_id,
            "task_type": "ocr",
            "status": "completed"
        }
        send_ack_message(ack_message, ack_channel)

        if ch.is_open:
            ch.basic_ack(delivery_tag=method.delivery_tag)
        else:
            logging.warning("Channel is closed. Cannot ack message.")

    except Exception as e:
        logging.error(f"Error in callback: {e}")
        if ch.is_open:
            ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
        else:
            logging.warning("Channel is closed. Cannot nack message.")

        ack_message = {
            "meeting_id": meeting_id,
            "task_id": task_id,
            "task_type": "ocr",
            "status": "failed"
        }
        send_ack_message(ack_message, ack_channel)


def send_ack_message(message, ack_channel):
    try:
        ack_channel.queue_declare(queue='orchestrator_ack_queue', durable=True)

        ack_channel.basic_publish(
            exchange='',
            routing_key='orchestrator_ack_queue',
            body=json.dumps(message),
            properties=pika.BasicProperties(
                content_type='application/json'
            )
        )
        logging.info(f"Acknowledgment sent: {message}")

    except Exception as e:
        logging.error(f"Failed to send acknowledgment message: {e}")


def main():
    credentials = pika.PlainCredentials('guest', 'guest')

    while True:
        try:
            logging.info("Connecting to RabbitMQ...")
            connection = pika.BlockingConnection(pika.ConnectionParameters(
                host='rabbitmq',
                port=5672,
                credentials=credentials,
                connection_attempts=5,
                retry_delay=5,
                socket_timeout=10,
                heartbeat=900
            ))
            channel = connection.channel()
            ack_channel = connection.channel()
            logging.info("Connected to RabbitMQ")

            channel.queue_declare(queue='ocr_queue', durable=True)
            logging.info("Queue declared.")

            channel.basic_qos(prefetch_count=1)
            channel.basic_consume(
                queue='ocr_queue',
                on_message_callback=lambda ch, method, properties, body: process_message(
                    ch, method, properties, body, ack_channel
                ),
                auto_ack=False,
                consumer_tag="ocr_consumer"
            )

            logging.info("Waiting for messages. To exit press CTRL+C")
            channel.start_consuming()

        except (pika.exceptions.ConnectionClosedByBroker,
                pika.exceptions.AMQPChannelError,
                pika.exceptions.AMQPConnectionError) as e:
            logging.error(f"Connection error: {e}. Retrying in 5 seconds...")
            time.sleep(5)

if __name__ == "__main__":
    main()