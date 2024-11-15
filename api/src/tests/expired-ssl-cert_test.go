package tests

import (
	"testing"
	"time"

	"infraops.dev/statuspage-core/utils"
)

func TestOnExpiredCert(t *testing.T) {
	var client = HttpClient()

	certInfos, err := utils.FetchCertInfo("expired.badssl.com:443")
	if err != nil {
		ApiV1(*client, "/admins-team/schommer-intern", map[string]interface{}{
			"reason":    err.Error(),
			"timestamp": time.Now().Format(time.RFC3339),
			"status":    "FATAL",
		})
		t.Fatalf("failed to fetch SSL certificate information.\nReason: %s", err)
	}

	for _, certInfo := range certInfos {
		if !certInfo.Valid {
			ApiV1(*client, "/admins-team/schommer-intern", map[string]interface{}{
				"reason":    "SSL certificate expired / IS NOT VALID",
				"timestamp": time.Now().Format(time.RFC3339),
				"status":    "FATAL",
			})
			t.Fatalf("SSL certificate expired / IS NOT VALID.\nReason: %s", err)
		}
	}

	ApiV1(*client, "/admins-team/schommer-intern", map[string]interface{}{
		"reason":    "No expired certificates found",
		"timestamp": time.Now().Format(time.RFC3339),
		"status":    "INFO",
	})
	t.Log("no expired certificates found")
}
