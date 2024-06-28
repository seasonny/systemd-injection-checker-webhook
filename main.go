package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validatePod(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var admissionReview admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		http.Error(w, "Error parsing admission review", http.StatusBadRequest)
		return
	}

	var pod corev1.Pod
	if err := json.Unmarshal(admissionReview.Request.Object.Raw, &pod); err != nil {
		http.Error(w, "Error parsing pod object", http.StatusBadRequest)
		return
	}

	allowed := true
	var reason string

	for key := range pod.Annotations {
		if strings.HasPrefix(key, "org.systemd.property.") {
			allowed = false
			reason = fmt.Sprintf("Annotation with prefix 'org.systemd.property.' is not allowed: %s", key)
			break
		}
	}

	response := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     admissionReview.Request.UID,
			Allowed: allowed,
		},
	}

	if !allowed {
		response.Response.Result = &metav1.Status{
			Message: reason,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	certFile := os.Getenv("CERT_FILE")
	keyFile := os.Getenv("KEY_FILE")

	if certFile == "" || keyFile == "" {
		panic("CERT_FILE and KEY_FILE environment variables must be set")
	}

	http.HandleFunc("/validate", validatePod)

	go func() {
		fmt.Println("Starting HTTP server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			panic(err)
		}
	}()

	fmt.Println("Starting HTTPS server on :8443")
	server := &http.Server{
		Addr: ":8443",
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
		panic(err)
	}
}
