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

type stringList []string

func (sl *stringList) String() string {
	return ""
}

func (sl *stringList) Set(s string) error {
	*sl = append(*sl, s)
	return nil
}

var (
	url           = flag.String("url", "", "MQTT URL")
	clientID      = flag.String("client-id", "mqtt-exporter", "MQTT client ID")
	topicPatterns = stringList{}
	listenAddr    = flag.String("addr", ":8080", "HTTP server listening address")

	topicGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mqtt_last_message",
			Help: "MQTT topic time",
		},
		[]string{topicLabel},
	)
)

func main() {
	flag.Var(&topicPatterns, "topic-pattern", "Subscribe to topic pattern (can be repeated)")
	flag.Parse()
	if *url == "" || len(topicPatterns) == 0 || (len(topicPatterns) == 1 && topicPatterns[0] == "") {
		fmt.Println("-url and -topic-pattern are both required")
		flag.Usage()
		os.Exit(1)
	}

	prometheus.MustRegister(topicGauge)

	mqtt.ERROR = log.New(os.Stderr, "[mqtt] ", log.LstdFlags)

	log.Println("Starting mqtt-exporter...")

	opts := mqtt.NewClientOptions().AddBroker(*url).SetClientID(*clientID)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetConnectionLostHandler(connectionLostHandler)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalln("Failed to connect to MQTT:", token.Error())
	}

	for _, topicPattern := range topicPatterns {
		if token := client.Subscribe(topicPattern, 0, messageHandler); token.Wait() && token.Error() != nil {
			client.Disconnect(0)
			log.Fatalln("Failed to subscribe to MQTT:", token.Error())
		}
	}

	log.Println("Listening on:", *listenAddr)
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
