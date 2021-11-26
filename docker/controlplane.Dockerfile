FROM golang:1.15-buster AS builder
ENV PROJECT_DIR /go/src/github.com/chez-shanpu/acar
COPY ./ ${PROJECT_DIR}/
WORKDIR ${PROJECT_DIR}/
RUN go build -o bin/controlplane ./cmd/controlplane


FROM ubuntu:focal
COPY --from=builder /go/src/github.com/chez-shanpu/acar/bin/controlplane /
ENTRYPOINT ["/controlplane"]
CMD ["--help"]
