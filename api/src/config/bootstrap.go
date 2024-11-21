package config

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var RootCAs *x509.CertPool
var DbConn *pgx.Conn
var AppKey string = "base64:qZCQSIA7VPk8Zxuc+lk/LJOeyoxTnU/hpesawf8gL2s="
var Clientset *kubernetes.Clientset

func certPool() {
	var err error
	RootCAs, err = x509.SystemCertPool()
	if err != nil || RootCAs == nil {
		RootCAs = x509.NewCertPool()
		log.Println("Using new cert pool.")
	}
}

func Bootstrap() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetOutput(os.Stdout)

	// Get KUBECONFIG from environment variable
	kubeconfig, exists := os.LookupEnv("KUBECONFIG")
	if exists && kubeconfig != "" {
		fmt.Println("⎈ Kubernetes-environment detected.")
	}

	// Build Kubernetes config
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(fmt.Errorf("failed to build Kubernetes config: %v", err))
	}

	dumpConfig(config)

	Clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Errorf("failed to create Kubernetes client: %v", err))
	}

	watchPods(Clientset, "sh-jenniferwalker")

	certPool()
}

func TCPMonitoring(host string, port string) (time.Duration, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 3*time.Second)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	elapsed := time.Since(start)
	return elapsed, nil
}

func retryTCPWithBackoff(host, port string, retries int) (time.Duration, error) {
	var lastErr error
	delay := 2 * time.Second
	for i := 0; i < retries; i++ {
		latency, err := TCPMonitoring(host, port)
		if err == nil {
			return latency, nil
		}
		lastErr = err
		log.Printf("Retry %d/%d for %s:%s failed: %v. Retrying in %v...\n", i+1, retries, host, port, err, delay)
		time.Sleep(delay)
		delay *= 2 // Exponential backoff
	}
	return 0, lastErr
}

// handlePodEvent processes a pod and performs TCP monitoring.
func handlePodEvent(pod *v1.Pod) {
	if pod.Status.PodIP == "" {
		log.Printf("Pod %s/%s has no IP yet, skipping\n", pod.Namespace, pod.Name)
		return
	}

	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		log.Printf("Skipping pod %s/%s with phase %s\n", pod.Namespace, pod.Name, pod.Status.Phase)
		return
	}

	if pod.Status.Phase != v1.PodRunning {
		log.Printf("Pod %s/%s is not in Running state, skipping\n", pod.Namespace, pod.Name)
		return
	}

	for _, container := range pod.Spec.Containers {
		for _, envVar := range container.Env {
			if envVar.Name == "CLIENT_DOMAIN" {
				clientDomain := envVar.Value
				fmt.Printf("Pod %s in namespace %s exposes CLIENT_DOMAIN: %s\n", pod.Name, pod.Namespace, clientDomain)
				latency, err := retryTCPWithBackoff(clientDomain, "443", 3)
				if err != nil {
					fmt.Printf("TCP lookup failed for domain %s: %v\n", clientDomain, err)
				} else {
					fmt.Printf("TCP lookup succeeded for domain %s. Latency: %v\n", clientDomain, latency)
				}
			}
		}
	}
}

func retryTCP(host, port string, retries int) (time.Duration, error) {
	var lastErr error
	for i := 0; i < retries; i++ {
		latency, err := TCPMonitoring(host, port)
		if err == nil {
			return latency, nil
		}
		lastErr = err
		time.Sleep(2 * time.Second) // Backoff
	}
	return 0, lastErr
}

func watchPods(clientset *kubernetes.Clientset, namespace string) {
	watcher, err := clientset.CoreV1().Pods(namespace).Watch(context.TODO(), metav1.ListOptions{
		LabelSelector: "workload-class=webstack-php",
	})
	if err != nil {
		log.Fatalf("Error watching pods in namespace %s: %v", namespace, err)
	}

	for event := range watcher.ResultChan() {
		pod, ok := event.Object.(*v1.Pod)
		if !ok {
			log.Println("Unexpected type")
			continue
		}

		switch event.Type {
		case watch.Added:
			log.Printf("Pod added: %s/%s\n", namespace, pod.Name)
			handlePodEvent(pod)
		case watch.Modified:
			log.Printf("Pod modified: %s/%s\n", namespace, pod.Name)
			handlePodEvent(pod)
		case watch.Deleted:
			log.Printf("Pod deleted: %s/%s\n", namespace, pod.Name)
		default:
			log.Printf("Unhandled event type: %v\n", event.Type)
		}
	}
}

// saveMetrics saves TCP lookup time metrics (replace with your own database or system).
func saveMetrics(namespace, podName, ip, port string, latency time.Duration) {
	// Replace this with your own storage logic (e.g., push to a database, time-series DB, etc.)
	log.Printf("Saving metrics - Namespace: %s, Pod: %s, IP: %s, Port: %s, Latency: %v\n",
		namespace, podName, ip, port, latency)
}

func dumpConfig(config *rest.Config) {
	// Define a custom struct to omit unsupported fields
	type ConfigDump struct {
		Host        string `json:"host"`
		APIPath     string `json:"apiPath,omitempty"`
		ContentType string `json:"contentType,omitempty"`
		BearerToken string `json:"bearerToken,omitempty"`
		// TLSClientConfig rest.TLSClientConfig `json:"tlsClientConfig"`
		UserAgent string  `json:"userAgent,omitempty"`
		QPS       float32 `json:"qps,omitempty"`
		Burst     int     `json:"burst,omitempty"`
		Timeout   int64   `json:"timeout,omitempty"`
	}

	// Populate the custom struct
	dump := ConfigDump{
		Host:        config.Host,
		APIPath:     config.APIPath,
		ContentType: config.ContentType,
		BearerToken: config.BearerToken,
		// TLSClientConfig: config.TLSClientConfig,
		UserAgent: config.UserAgent,
		QPS:       config.QPS,
		Burst:     config.Burst,
		Timeout:   config.Timeout.Milliseconds(),
	}

	// Marshal the custom struct to JSON
	configJSON, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		panic(fmt.Errorf("Failed to marshal Kubernetes config: %v", err))
	}
	fmt.Printf("Kubernetes Config: %s\n", string(configJSON))
}
