FROM golang:1.26.8 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/prometheus-k8s-cert-exporter \
    ./cmd/exporter

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=builder \
    --chown=nonroot:nonroot \
    /out/prometheus-k8s-cert-exporter \
    /prometheus-k8s-cert-exporter

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/prometheus-k8s-cert-exporter"]