package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"iammati/statuspage/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListNamespaces() string {
	namespaces, err := fetchNamespaces()
	if err != nil {
		fmt.Printf("Error fetching namespaces: %v\n", err)
		// Return an error response as a JSON string
		errorResponse := map[string]string{"error": "Failed to fetch namespaces"}
		errorJSON, _ := json.Marshal(errorResponse) // Ignoring error since it's simple JSON
		return string(errorJSON)
	}

	responseData := map[string]interface{}{
		"api":        "k8s/namespaces/list",
		"namespaces": namespaces,
	}

	// Convert the response data to JSON
	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		fmt.Printf("Error marshalling namespaces response: %v\n", err)
		// Return an error response as a JSON string
		errorResponse := map[string]string{"error": "Failed to generate response"}
		errorJSON, _ := json.Marshal(errorResponse) // Ignoring error since it's simple JSON
		return string(errorJSON)
	}

	return string(responseJSON)
}

func fetchNamespaces() (string, error) {
	namespaces, err := config.Clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to fetch namespaces: %v", err)
	}

	if len(namespaces.Items) == 0 {
		return "[]", nil
	}

	var namespaceNames []string
	for _, ns := range namespaces.Items {
		namespaceNames = append(namespaceNames, ns.Name)
	}

	jsonData, err := json.Marshal(namespaceNames)
	if err != nil {
		return "", fmt.Errorf("failed to convert namespace names to JSON: %v", err)
	}

	return string(jsonData), nil
}
