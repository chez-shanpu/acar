FROM golang:1.15-buster AS builder
ENV PROJECT_DIR /go/src/github.com/chez-shanpu/acar
COPY ./ ${PROJECT_DIR}/
WORKDIR ${PROJECT_DIR}/
RUN go build -o bin/monitoring-server ./cmd/monitoring-server


FROM gcr.io/distroless/base-debian10
COPY --from=builder /go/src/github.com/chez-shanpu/traffic-generator/bin/monitoring-server /
ENTRYPOINT ["/monitoring-server"]
CMD ["--help"]
