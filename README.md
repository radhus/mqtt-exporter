# mqtt-exporter

Publishes Prometheus metrics for last received message on a given MQTT topic.

## Usage

Docker image is built and published to [`radhus/mqtt-exporter` on Docker Hub](https://hub.docker.com/r/radhus/mqtt-exporter).

```bash
docker run -p 8080:8080 radhus/mqtt-exporter:latest -url mqtt-server.domain:1883 -topic-pattern 'topic/+/pattern'
```

Then scrape `http://localhost:8080/metrics` to get:

```
# HELP mqtt_last_message MQTT topic time
# TYPE mqtt_last_message gauge
mqtt_last_message{topic="topic/11/pattern"} 1.5825474109081864e+09
mqtt_last_message{topic="topic/12/pattern"} 1.5825474109117537e+09
```

Simple Prometheus alert may look like:

```yaml
- alert: NoIOTStateUpdate
  expr: (time() - mqtt_last_message{topic="topic/11/pattern"}) / 60 > 5
  for: 5m
  annotations:
    description: "IOT state has not updated in {{ $value }} minutes!"
    summary: "IOT state stopped updating"
```
