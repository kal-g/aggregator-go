FROM golang:1.15.0

RUN apt-get update && apt-get install -y bc
COPY aggregator-go aggregator-go
WORKDIR aggregator-go

ENTRYPOINT make all && NODE_NAME=agg ./scripts/e2e_test.sh && make benchmark