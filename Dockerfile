#################
# Base image
#################
FROM alpine:3.12 as echo-kafka-base

USER root

RUN addgroup -g 10001 echo-kafka && \
    adduser --disabled-password --system --gecos "" --home "/home/echo-kafka" --shell "/sbin/nologin" --uid 10001 echo-kafka && \
    mkdir -p "/home/echo-kafka" && \
    chown echo-kafka:0 /home/echo-kafka && \
    chmod g=u /home/echo-kafka && \
    chmod g=u /etc/passwd
RUN apk add --update --no-cache alpine-sdk curl

ENV USER=echo-kafka
USER 10001
WORKDIR /home/echo-kafka

#################
# Builder image
#################
FROM golang:1.16-alpine AS echo-kafka-builder
RUN apk add --update --no-cache alpine-sdk
WORKDIR /app
COPY . .
RUN make build

#################
# Final image
#################
FROM echo-kafka-base

COPY --from=echo-kafka-builder /app/bin/echo-kafka /usr/local/bin

# Command to run the executable
ENTRYPOINT ["echo-kafka"]
