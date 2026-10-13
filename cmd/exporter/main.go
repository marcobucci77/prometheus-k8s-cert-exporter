package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/marcobucci77/prometheus-k8s-cert-exporter/internal/collector"
	"github.com/marcobucci77/prometheus-k8s-cert-exporter/internal/kubeclient"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	clientset, configSource, err := kubeclient.NewClient()
	if err != nil {
		log.Fatalf("failed to create Kubernetes client: %v", err)
	}

	log.Printf("Kubernetes client configured using %s", configSource)

	rawNamespaces := os.Getenv("WATCH_NAMESPACES")

	if rawNamespaces == "" {
		log.Fatal("WATCH_NAMESPACES environment variable is required")
	}

	var namespaces []string

	for _, namespace := range strings.Split(rawNamespaces, ",") {
		namespace = strings.TrimSpace(namespace)

		if namespace != "" {
			namespaces = append(namespaces, namespace)
		}
	}

	if len(namespaces) == 0 {
		log.Fatal("WATCH_NAMESPACES does not contain any valid namespace")
	}

	log.Printf(
		"Monitoring Kubernetes namespaces: %s",
		strings.Join(namespaces, ", "),
	)

	certificateCollector := collector.NewCertificateCollector(
		clientset,
		namespaces,
	)

	prometheus.MustRegister(certificateCollector)

	http.Handle("/metrics", promhttp.Handler())

	log.Println("prometheus-k8s-cert-exporter listening on :8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
