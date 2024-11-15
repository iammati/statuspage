package tests

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"infraops.dev/statuspage-core/config"
	"infraops.dev/statuspage-core/utils"
)

func CertPool() *x509.CertPool {
	caCertPool, err := utils.LoadCertsFromDir("/usr/local/share/ca-certificates")
	if err != nil {
		log.Fatalf("failed to load custom CA certificates.\nReason: %s", err)
	}
	return caCertPool
}

func HttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            CertPool(),
				InsecureSkipVerify: true,
			},
		},
	}
}

func ApiV1(client http.Client, url string, data map[string]interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", "https://ddev-app-web/api/v1"+url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("X-App-Key", config.AppKey)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API/v1 returned status code %d", resp.StatusCode)
	}

	return nil
}
