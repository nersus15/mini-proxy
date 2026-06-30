## BUILD STAGE
FROM golang:1.25-trixie

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    librdkafka-dev \
    librdkafka1 \
    pkg-config \
    vim \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY ./main /usr/local/bin/webcore
COPY ./config.yaml ./config.yaml
COPY ./webcore ./webcore
COPY ./db ./db


# Siapkan folder untuk sqlite (opsional)
RUN mkdir -p /var/miniproxy


CMD ["webcore","proxy"]

