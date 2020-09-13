FROM golang:1.15.0

RUN apt-get update && apt-get install -y librocksdb-dev bc
COPY aggregator-go aggregator-go
WORKDIR aggregator-go
RUN make

EXPOSE 50051

ENTRYPOINT ./bin/aggregator "bin/rocksdb_storage"