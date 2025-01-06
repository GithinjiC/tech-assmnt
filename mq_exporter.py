from prometheus_client import start_http_server, Gauge
import requests
import time
import argparse

total_messages_gauge = Gauge(
    'rabbitmq_queue_messages_total',
    'Total number of messages in queues.'
    ['vhost', 'queue']
)

messages_ready_gauge = Gauge(
    'rabbitmq_queue_messages_ready',
    'Number of ready messages in queues.'
    ['vhost', 'queue']
)

messages_unacknowledged_gauge = Gauge(
    'rabbitmq_queue_messages_total',
    'Number of unacknowledged messages in queues.'
    ['vhost', 'queue']
)

def fetch_rabbitmq_metrics(api_url, username, password):
    try:
        response = requests.get(f'{api_url}/queues', auth=(username, password))
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f"Error fetching RabbitMQ metrics: {e}")
        return []
    
def update_metrics(data):
    for queue in data:
        vhost = queue['vhost']
        name = queue['name']
        total_messages = queue.get('messages', 0)
        messages_ready = queue.get('message_ready', 0)
        messages_unacknowledged = queue.get('messages_unacknowledged', 0)

        total_messages_gauge.labels(vhost=vhost, queue=name).set(total_messages)
        messages_ready_gauge.labels(vhost=vhost, queue=name).set(messages_ready)
        messages_unacknowledged_gauge.labels(vhost=vhost, queue=name).set(messages_unacknowledged)

def main():
    parser = argparse.ArgumentParser(description='RabbitMQ Prometheus Exporter')
    parser.add_argument('--api-url', required=True, help='RabbitMQ HTTP APU URL')
    parser.add_argument('--username', required=True, help='RabbitMQ username')
    parser.add_argument('--password', required=True, help='RabbitMQ password')
    parser.add_argument('--port', type=int, default=8000, help='Exposed port to Prometheus')
    parser.add_argument('--interval', type=int, default=15, help='Interval to seconds to fetch metrics')

    args = parser.parse_args()

    start_http_server(args.port)
    print(f"Starting Prometheus exporter on port {args.port}...")

    while True:
        data = fetch_rabbitmq_metrics(args.api_url, args.username, args.password)
        update_metrics(data)
        time.sleep(args.interval)

if __name__ == '__main__':
    main()