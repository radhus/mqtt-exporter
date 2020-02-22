FROM golang:1.13 AS builder

WORKDIR /go/src/app
ADD . /go/src/app
RUN go build -o /go/bin/mqtt-exporter

FROM gcr.io/distroless/base
COPY --from=builder /go/bin/mqtt-exporter /
CMD ["/mqtt-exporter"]
