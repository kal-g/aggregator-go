FROM golang:1.15.0

RUN apt-get update && apt-get install -y librocksdb-dev bc

RUN git clone https://github.com/kal-g/aggregator-go.git

WORKDIR /home/aggregator-go

CMD make benchmark