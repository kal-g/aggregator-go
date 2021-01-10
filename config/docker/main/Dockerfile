FROM golang:1.15.0

COPY aggregator-go aggregator-go
WORKDIR aggregator-go
RUN mkdir -p bin && go build -o bin/aggregator cmd/aggregator/main.go

EXPOSE 50051

ENTRYPOINT ["./bin/aggregator", "--config", "config/aggregator_configs/global", "--config", "config/aggregator_configs/test"]