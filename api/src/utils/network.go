package utils

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"infraops.dev/statuspage-core/config"
)

type Metrics struct {
	DnsResolutionTime time.Duration
	TcpConnectionTime time.Duration
	TlsConnectionTime time.Duration
	HttpTime          time.Duration
	Reachable         bool
	StatusCode        int
	Error             error
}

func HostMetrics(hostname string, path string) (Metrics, error) {
	var metrics Metrics
	host, port, err := net.SplitHostPort(hostname)
	if err != nil {
		host = hostname
		port = "443"
	}

	// DNS Resolution
	start := time.Now()
	ips, err := net.LookupIP(host)
	metrics.DnsResolutionTime = time.Since(start)
	if err != nil || len(ips) == 0 {
		metrics.Reachable = false
		return metrics, fmt.Errorf("DNS resolution failed for %s", host)
	}

	// TCP Connection
	resolvedHost := net.JoinHostPort(ips[0].String(), port)
	conn, err := net.DialTimeout("tcp", resolvedHost, 5*time.Second)
	metrics.TcpConnectionTime = time.Since(start) - metrics.DnsResolutionTime
	if err != nil {
		metrics.Reachable = false
		return metrics, fmt.Errorf("TCP connection failed for %s", resolvedHost)
	}
	// Do not close the TCP connection here; we need it for the TLS handshake

	// Load custom CA certificates
	caCertPool, err := LoadCertsFromDir("/usr/local/share/ca-certificates")
	if err != nil {
		metrics.Reachable = false
		conn.Close() // Close the connection in case of error
		return metrics, fmt.Errorf("failed to load custom CA certificates.\nReason: %s", err)
	}

	// TLS Handshake
	tlsStart := time.Now()
	tlsConn := tls.Client(conn, &tls.Config{
		ServerName: host,
		RootCAs:    caCertPool,
	})
	err = tlsConn.Handshake()
	metrics.TlsConnectionTime = time.Since(tlsStart)
	if err != nil {
		metrics.Reachable = false
		conn.Close() // Close the connection in case of error
		return metrics, fmt.Errorf("TLS handshake failed for %s: %v", resolvedHost, err)
	}
	defer tlsConn.Close() // Close the TLS connection after successful handshake

	// Create HTTP client with custom transport
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
		Timeout: 5 * time.Second,
	}

	// HTTP Request
	httpStart := time.Now()
	URL := "https://" + host + path
	response, err := client.Get(URL)
	metrics.HttpTime = time.Since(httpStart)
	if err != nil || response.StatusCode >= 400 {
		metrics.Reachable = false
		metrics.Error = fmt.Errorf("HTTP request failed for %s.\nReason: %s", "'"+URL+"'", err)
	}
	defer response.Body.Close()

	if (response.StatusCode >= 200 && response.StatusCode < 400) || response.StatusCode == 0 {
		metrics.Reachable = true
	}
	metrics.StatusCode = response.StatusCode

	return metrics, nil
}

func FetchCertInfo(host string) ([]CertInfo, error) {
	hostName := strings.Split(host, ":")[0]
	conn, err := tls.Dial("tcp", host, &tls.Config{
		RootCAs:    config.RootCAs,
		ServerName: hostName,
	})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var certInfos []CertInfo
	for _, cert := range conn.ConnectionState().PeerCertificates {
		wildcardNames := []string{}
		for _, dnsName := range cert.DNSNames {
			if strings.Contains(dnsName, "*") {
				wildcardNames = append(wildcardNames, dnsName)
			}
		}
		certInfos = append(certInfos, CertInfo{
			Issuer:        cert.Issuer.String(),
			Subject:       cert.Subject.String(),
			Expiration:    cert.NotAfter.Format(time.RFC3339),
			Valid:         cert.NotAfter.After(time.Now()),
			WildcardNames: wildcardNames,
		})
	}
	return certInfos, nil
}

func HttpError(w http.ResponseWriter, errorMsg string, code int) {
	http.Error(w, errorMsg, code)
}

func JsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal JSON: %v", err)
		return
	}
	_, writeErr := w.Write(jsonData)
	if writeErr != nil {
		log.Printf("Failed to write JSON response: %v", writeErr)
	}
}

func wildcardLevel(dnsNames []string) int {
	maxLevel := 0
	for _, name := range dnsNames {
		if strings.Count(name, "*.") > maxLevel {
			maxLevel = strings.Count(name, "*.")
		}
	}
	return maxLevel
}

type CertInfo struct {
	Issuer        string   `json:"issuer"`
	Subject       string   `json:"subject"`
	Expiration    string   `json:"expiration"`
	Valid         bool     `json:"valid"`
	WildcardNames []string `json:"wildcardNames"`
}
