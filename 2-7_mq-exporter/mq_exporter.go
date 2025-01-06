// ./mq_exporter --port 8000 --interval 15
// curl http://localhost:8000/metrics

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type QueueMetrics struct {
	Name               string  `json:"name"`
	Vhost              string  `json:"vhost"`
	Messages           float64 `json:"messages"`
	MessagesReady      float64 `json:"messages_ready"`
	MessagesUnacknowledged float64 `json:"messages_unacknowledged"`
}

type Exporter struct {
	rabbitMQHost     string
	rabbitMQUser     string
	rabbitMQPassword string
	client           *http.Client

	messages           *prometheus.Desc
	messagesReady      *prometheus.Desc
	messagesUnacknowledged *prometheus.Desc
}

func NewExporter(host, user, password string) *Exporter {
	return &Exporter{
		rabbitMQHost:     host,
		rabbitMQUser:     user,
		rabbitMQPassword: password,
		client:           &http.Client{Timeout: 10 * time.Second},

		messages: prometheus.NewDesc(
			"rabbitmq_individual_queue_messages",
			"Total number of messages in RabbitMQ queue",
			[]string{"host", "vhost", "name"},
			nil,
		),
		messagesReady: prometheus.NewDesc(
			"rabbitmq_individual_queue_messages_ready",
			"Total number of ready messages in RabbitMQ queue",
			[]string{"host", "vhost", "name"},
			nil,
		),
		messagesUnacknowledged: prometheus.NewDesc(
			"rabbitmq_individual_queue_messages_unacknowledged",
			"Total number of unacknowledged messages in RabbitMQ queue",
			[]string{"host", "vhost", "name"},
			nil,
		),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.messages
	ch <- e.messagesReady
	ch <- e.messagesUnacknowledged
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	url := fmt.Sprintf("http://%s/api/queues", e.rabbitMQHost)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request: %v", err)
		return
	}
	req.SetBasicAuth(e.rabbitMQUser, e.rabbitMQPassword)

	resp, err := e.client.Do(req)
	if err != nil {
		log.Printf("Failed to fetch data from RabbitMQ API: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected status code from RabbitMQ API: %d", resp.StatusCode)
		return
	}

	var queues []QueueMetrics
	if err := json.NewDecoder(resp.Body).Decode(&queues); err != nil {
		log.Printf("Failed to decode RabbitMQ API response: %v", err)
		return
	}

	host := e.rabbitMQHost
	for _, queue := range queues {
		ch <- prometheus.MustNewConstMetric(
			e.messages,
			prometheus.GaugeValue,
			queue.Messages,
			host, queue.Vhost, queue.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			e.messagesReady,
			prometheus.GaugeValue,
			queue.MessagesReady,
			host, queue.Vhost, queue.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			e.messagesUnacknowledged,
			prometheus.GaugeValue,
			queue.MessagesUnacknowledged,
			host, queue.Vhost, queue.Name,
		)
	}
}

func main() {
	rabbitMQHost := os.Getenv("RABBITMQ_HOST")
	rabbitMQUser := os.Getenv("RABBITMQ_USER")
	rabbitMQPassword := os.Getenv("RABBITMQ_PASSWORD")

	if rabbitMQHost == "" || rabbitMQUser == "" || rabbitMQPassword == "" {
		log.Fatal("Environment variables RABBITMQ_HOST, RABBITMQ_USER, and RABBITMQ_PASSWORD must be set")
	}

	port := flag.String("port", "8000", "Port to expose metrics")
	interval := flag.Int("interval", 15, "Interval to scrape RabbitMQ metrics (in seconds)")
	flag.Parse()

	exporter := NewExporter(rabbitMQHost, rabbitMQUser, rabbitMQPassword)
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())
	serverAddr := fmt.Sprintf(":%s", *port)

	log.Printf("Starting RabbitMQ exporter on %s (scraping every %d seconds)", serverAddr, *interval)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}
