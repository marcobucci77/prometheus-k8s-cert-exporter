# prometheus-k8s-cert-exporter

Prometheus exporter for monitoring X.509 certificates stored in Kubernetes TLS Secrets.

## Features

- Discovers `kubernetes.io/tls` Secrets across Kubernetes namespaces
- Parses PEM/X.509 certificates
- Exposes certificate expiration timestamps
- Exposes certificate NotBefore timestamps
- Reports certificate parsing errors
- Reports current certificate validity
- Exposes certificate metadata such as Common Name, Issuer and CA status
- Supports in-cluster Kubernetes authentication
- Prometheus-compatible `/metrics` endpoint

## Metrics

- `k8s_cert_exporter_certificates_total`
- `k8s_cert_exporter_certificate_expiry_timestamp_seconds`
- `k8s_cert_exporter_certificate_not_before_timestamp_seconds`
- `k8s_cert_exporter_certificate_info`
- `k8s_cert_exporter_certificate_parse_success`
- `k8s_cert_exporter_certificate_valid`
- `k8s_cert_exporter_scrape_success`

## License

Apache License 2.0