FROM golang:1.15.0

RUN apt-get update && apt-get install -y bc
COPY aggregator-go aggregator-go
WORKDIR aggregator-go
RUN make

EXPOSE 50051

ENTRYPOINT ["./bin/aggregator"]