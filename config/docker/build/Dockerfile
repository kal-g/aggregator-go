FROM golang:1.15.0

RUN apt-get update && \
    apt-get install -y \
    librocksdb-dev \ 
    bc \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common \
    qemu binfmt-support qemu-user-static && \
    curl -fsSL https://download.docker.com/linux/debian/gpg | apt-key add - && \
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/debian $(lsb_release -cs) stable" && \
    apt-get update && \
    apt-get install -y qemu-user-static docker-ce-cli

RUN mkdir -vp ~/.docker/cli-plugins/ && \
    curl --silent -L --output ~/.docker/cli-plugins/docker-buildx https://github.com/docker/buildx/releases/download/v0.3.1/buildx-v0.3.1.linux-amd64 && \
    chmod a+x ~/.docker/cli-plugins/docker-buildx