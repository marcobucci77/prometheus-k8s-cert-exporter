package collector

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type CertificateCollector struct {
	client kubernetes.Interface

	certificatesTotal       *prometheus.Desc
	certificateExpiry       *prometheus.Desc
	certificateNotBefore    *prometheus.Desc
	certificateInfo         *prometheus.Desc
	certificateParseSuccess *prometheus.Desc
	certificateValid        *prometheus.Desc
	scrapeSuccess           *prometheus.Desc
}

func NewCertificateCollector(
	client kubernetes.Interface,
) *CertificateCollector {

	return &CertificateCollector{
		client: client,
		certificateValid: prometheus.NewDesc(
			"k8s_cert_exporter_certificate_valid",
			"Whether the Kubernetes TLS certificate is currently within its validity period.",
			[]string{"namespace", "secret"},
			nil,
		),

		certificatesTotal: prometheus.NewDesc(
			"k8s_cert_exporter_certificates_total",
			"Total number of Kubernetes TLS certificate Secrets discovered.",
			nil,
			nil,
		),

		certificateExpiry: prometheus.NewDesc(
			"k8s_cert_exporter_certificate_expiry_timestamp_seconds",
			"Unix timestamp when the Kubernetes TLS certificate expires.",
			[]string{"namespace", "secret"},
			nil,
		),

		certificateNotBefore: prometheus.NewDesc(
			"k8s_cert_exporter_certificate_not_before_timestamp_seconds",
			"Unix timestamp when the Kubernetes TLS certificate becomes valid.",
			[]string{"namespace", "secret"},
			nil,
		),

		certificateInfo: prometheus.NewDesc(
			"k8s_cert_exporter_certificate_info",
			"Information about a Kubernetes TLS certificate.",
			[]string{
				"namespace",
				"secret",
				"common_name",
				"issuer",
				"is_ca",
			},
			nil,
		),

		certificateParseSuccess: prometheus.NewDesc(
			"k8s_cert_exporter_certificate_parse_success",
			"Whether the Kubernetes TLS certificate was parsed successfully.",
			[]string{"namespace", "secret"},
			nil,
		),

		scrapeSuccess: prometheus.NewDesc(
			"k8s_cert_exporter_scrape_success",
			"Whether collecting Kubernetes certificate information succeeded.",
			nil,
			nil,
		),
	}
}

func (c *CertificateCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.certificatesTotal
	ch <- c.certificateExpiry
	ch <- c.certificateNotBefore
	ch <- c.certificateInfo
	ch <- c.certificateParseSuccess
	ch <- c.certificateValid
	ch <- c.scrapeSuccess
}

func (c *CertificateCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	secrets, err := c.client.
		CoreV1().
		Secrets("").
		List(ctx, metav1.ListOptions{})

	if err != nil {
		log.Printf("failed to list Kubernetes Secrets: %v", err)

		ch <- prometheus.MustNewConstMetric(
			c.scrapeSuccess,
			prometheus.GaugeValue,
			0,
		)

		return
	}

	total := 0
	now := time.Now()

	for _, secret := range secrets.Items {
		if secret.Type != corev1.SecretTypeTLS {
			continue
		}

		total++

		certificateData, ok := secret.Data[corev1.TLSCertKey]
		if !ok {
			log.Printf(
				"Secret %s/%s does not contain %s",
				secret.Namespace,
				secret.Name,
				corev1.TLSCertKey,
			)

			ch <- prometheus.MustNewConstMetric(
				c.certificateParseSuccess,
				prometheus.GaugeValue,
				0,
				secret.Namespace,
				secret.Name,
			)

			continue
		}

		if len(certificateData) == 0 {
			log.Printf(
				"Secret %s/%s contains an empty %s",
				secret.Namespace,
				secret.Name,
				corev1.TLSCertKey,
			)

			ch <- prometheus.MustNewConstMetric(
				c.certificateParseSuccess,
				prometheus.GaugeValue,
				0,
				secret.Namespace,
				secret.Name,
			)

			continue
		}

		cert, err := parseCertificate(certificateData)
		if err != nil {
			log.Printf(
				"failed to parse certificate %s/%s: %v",
				secret.Namespace,
				secret.Name,
				err,
			)

			ch <- prometheus.MustNewConstMetric(
				c.certificateParseSuccess,
				prometheus.GaugeValue,
				0,
				secret.Namespace,
				secret.Name,
			)

			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.certificateExpiry,
			prometheus.GaugeValue,
			float64(cert.NotAfter.Unix()),
			secret.Namespace,
			secret.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.certificateNotBefore,
			prometheus.GaugeValue,
			float64(cert.NotBefore.Unix()),
			secret.Namespace,
			secret.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.certificateInfo,
			prometheus.GaugeValue,
			1,
			secret.Namespace,
			secret.Name,
			cert.Subject.CommonName,
			cert.Issuer.CommonName,
			strconv.FormatBool(cert.IsCA),
		)

		ch <- prometheus.MustNewConstMetric(
			c.certificateParseSuccess,
			prometheus.GaugeValue,
			1,
			secret.Namespace,
			secret.Name,
		)

		valid := 0.0

		if !now.Before(cert.NotBefore) && now.Before(cert.NotAfter) {
			valid = 1
		}

		ch <- prometheus.MustNewConstMetric(
			c.certificateValid,
			prometheus.GaugeValue,
			valid,
			secret.Namespace,
			secret.Name,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.certificatesTotal,
		prometheus.GaugeValue,
		float64(total),
	)

	ch <- prometheus.MustNewConstMetric(
		c.scrapeSuccess,
		prometheus.GaugeValue,
		1,
	)
}

func parseCertificate(data []byte) (*x509.Certificate, error) {
	for len(data) > 0 {
		block, rest := pem.Decode(data)

		if block == nil {
			return nil, fmt.Errorf("unable to decode PEM data")
		}

		data = rest

		if block.Type != "CERTIFICATE" {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to parse X.509 certificate: %w",
				err,
			)
		}

		return cert, nil
	}

	return nil, fmt.Errorf("no certificate found in PEM data")
}
