FROM golang:1.15.0

RUN apt-get update && apt-get install -y bc
COPY aggregator-go aggregator-go
WORKDIR aggregator-go

ENTRYPOINT make e2e_test && make benchmark