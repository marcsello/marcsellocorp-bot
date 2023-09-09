FROM golang:1.19-alpine3.17 as builder

COPY . /src/
WORKDIR /src

RUN apk add --no-cache make=4.3-r1 && make -j "$(nproc)"

FROM alpine:3.17

# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates

COPY --from=builder /src/main /app/main

EXPOSE 8080 8081

ENTRYPOINT [ "/app/main" ]