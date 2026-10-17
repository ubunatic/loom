FROM golang:1.25.0-bookworm

WORKDIR /src
ENV GOWORK=off

RUN apt-get update && apt-get install -y --no-install-recommends python3 && rm -rf /var/lib/apt/lists/*

COPY Makefile ./
RUN make lint-tools

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY scripts/container-test.sh /usr/local/bin/container-test
RUN chmod 0755 /usr/local/bin/container-test

ENTRYPOINT ["/usr/local/bin/container-test"]
