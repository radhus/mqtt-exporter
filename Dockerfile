FROM golang:1.21 AS builder

WORKDIR /go/src/app
ADD go.mod go.sum /go/src/app/
RUN go mod download

ADD . /go/src/app
RUN go build -o /go/bin/mqtt-exporter

FROM gcr.io/distroless/base
COPY --from=builder /go/bin/mqtt-exporter /
CMD ["/mqtt-exporter"]
