package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"iammati/statuspage/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListPods(namespace string) string {
	if namespace == "" {
		fmt.Println("Namespace is required")
		// Return an error response as a JSON string
		errorResponse := map[string]string{"error": "Namespace is mandatory"}
		errorJSON, _ := json.Marshal(errorResponse) // Ignoring error since it's simple JSON
		return string(errorJSON)
	}

	pods, err := fetchPods(namespace)
	if err != nil {
		fmt.Printf("Error fetching pods in namespace '%s': %v\n", namespace, err)
		// Return an error response as a JSON string
		errorResponse := map[string]string{"error": fmt.Sprintf("Failed to fetch pods in namespace '%s'", namespace)}
		errorJSON, _ := json.Marshal(errorResponse) // Ignoring error since it's simple JSON
		return string(errorJSON)
	}

	responseData := map[string]interface{}{
		"api":       "k8s/pods/list",
		"namespace": namespace,
		"pods":      pods,
	}

	// Convert the response data to JSON
	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		fmt.Printf("Error marshalling pods response: %v\n", err)
		// Return an error response as a JSON string
		errorResponse := map[string]string{"error": "Failed to generate response"}
		errorJSON, _ := json.Marshal(errorResponse) // Ignoring error since it's simple JSON
		return string(errorJSON)
	}

	return string(responseJSON)
}

func fetchPods(namespace string) (string, error) {
	pods, err := config.Clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to fetch pods in namespace '%s': %v", namespace, err)
	}

	if len(pods.Items) == 0 {
		return "[]", nil
	}

	var podNames []string
	for _, p := range pods.Items {
		podNames = append(podNames, p.Name)
	}

	jsonData, err := json.Marshal(podNames)
	if err != nil {
		return "", fmt.Errorf("failed to convert pod names to JSON: %v", err)
	}

	return string(jsonData), nil
}
