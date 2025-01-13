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

def process_message(ch, method, properties, body):
    logging.info("Callback triggered. Received message...")
    try:
        message = json.loads(body)
        logging.info(f"Message received: {message}")
        file_path = message.get("file_path")
        meeting_id = message.get("meeting_id")

        if not file_path or not meeting_id:
            logging.warning("Invalid message format. Skipping...")
            if ch.is_open:
                ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
            else:
                logging.warning("Channel is closed. Cannot ack message.")
            return

        logging.info(f"Received task for file: {file_path}, meeting ID: {meeting_id}")

        ocr_processing.ocr_image(file_path=file_path, meeting_id=meeting_id)

        if ch.is_open:
            ch.basic_ack(delivery_tag=method.delivery_tag)
        else:
            logging.warning("Channel is closed. Cannot ack message.")

    except Exception as e:
        logging.error(f"Error in callback: {e}")
        if ch.is_open:
            ch.basic_nack(delivery_tag=method.delivery_tag, requeue=True)
        else:
            logging.warning("Channel is closed. Cannot ack message.")

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
            logging.info("Connected to RabbitMQ")

            channel.queue_declare(queue='ocr_queue', durable=True)
            logging.info("Queue declared.")

            channel.basic_qos(prefetch_count=1)
            channel.basic_consume(queue='ocr_queue', on_message_callback=process_message, auto_ack=False, consumer_tag="ocr_consumer")

            logging.info("Waiting for messages. To exit press CTRL+C")
            channel.start_consuming()

        except (pika.exceptions.ConnectionClosedByBroker,
                pika.exceptions.AMQPChannelError,
                pika.exceptions.AMQPConnectionError) as e:
            logging.error(f"Connection error: {e}. Retrying in 5 seconds...")
            time.sleep(5)

if __name__ == "__main__":
    main()