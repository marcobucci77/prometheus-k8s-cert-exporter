package main

import (
	"log"
	"net/http"

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

	certificateCollector := collector.NewCertificateCollector(clientset)

	prometheus.MustRegister(certificateCollector)

	http.Handle("/metrics", promhttp.Handler())

	log.Println("prometheus-k8s-cert-exporter listening on :8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
