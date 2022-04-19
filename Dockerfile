
FROM golang:1.17-alpine as builder
RUN apk add make binutils
COPY / /work
WORKDIR /work
RUN make metal-exporter

FROM alpine:3.15
COPY --from=builder /work/bin/metal-exporter /metal-exporter
USER root
ENTRYPOINT ["/metal-exporter"]

EXPOSE 9080
