package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const topicLabel = "topic"

var (
	url          = flag.String("url", "", "MQTT URL")
	clientID     = flag.String("client-id", "mqtt-exporter", "MQTT client ID")
	topicPattern = flag.String("topic-pattern", "", "Subscribe to topic pattern")
	listenAddr   = flag.String("addr", ":8080", "HTTP server listening address")

	topicGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mqtt_last_message",
			Help: "MQTT topic time",
		},
		[]string{topicLabel},
	)
)

func main() {
	flag.Parse()
	if *url == "" || *topicPattern == "" {
		fmt.Println("-url and -topic-pattern are both required")
		flag.Usage()
		os.Exit(1)
	}

	prometheus.MustRegister(topicGauge)

	mqtt.ERROR = log.New(os.Stderr, "[mqtt] ", log.LstdFlags)

	opts := mqtt.NewClientOptions().AddBroker(*url).SetClientID(*clientID)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetConnectionLostHandler(connectionLostHandler)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalln("Failed to connect to MQTT:", token.Error())
	}

	if token := client.Subscribe(*topicPattern, 0, messageHandler); token.Wait() && token.Error() != nil {
		client.Disconnect(0)
		log.Fatalln("Failed to subscribe to MQTT:", token.Error())
	}

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(*listenAddr, nil)
	if err != nil {
		log.Fatalln("Server exited:", err)
	}
}

func messageHandler(_ mqtt.Client, message mqtt.Message) {
	if message.Retained() {
		return
	}

	topicGauge.With(prometheus.Labels{topicLabel: message.Topic()}).SetToCurrentTime()
}

func connectionLostHandler(_ mqtt.Client, err error) {
	log.Fatalln("Connection lost, terminating:", err)
}
