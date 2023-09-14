FROM golang:1.21-alpine3.18 as builder

COPY . /src/
WORKDIR /src

RUN apk add --no-cache make=4.3-r1 && make -j "$(nproc)"

FROM alpine:3.18

# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates

COPY --from=builder /src/main /app/main

EXPOSE 8080 8081

ENTRYPOINT [ "/app/main" ]