package utils

import (
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func LoadCertsFromDir(certDir string) (*x509.CertPool, error) {
	// Try to get the system cert pool, fall back to an empty pool on error
	certPool, err := x509.SystemCertPool()
	if err != nil || certPool == nil {
		certPool = x509.NewCertPool()
		log.Println("Using new cert pool.")
	}

	// Read all files from the specified directory
	files, err := os.ReadDir(certDir)
	if err != nil {
		return nil, fmt.Errorf("reading cert directory: %v", err)
	}

	// Try to append each file to the cert pool
	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			continue
		}

		certPath := filepath.Join(certDir, fileInfo.Name())
		certBytes, err := os.ReadFile(certPath)
		if err != nil {
			log.Printf("Failed to read %s: %v", certPath, err)
			continue // Log the error and move on to the next file
		}

		if ok := certPool.AppendCertsFromPEM(certBytes); !ok {
			log.Printf("Failed to append %s to cert pool", certPath)
		}
	}

	return certPool, nil
}
