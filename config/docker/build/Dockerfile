FROM golang:1.15.0

RUN apt-get update && apt-get install -y librocksdb-dev \
    bc \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -