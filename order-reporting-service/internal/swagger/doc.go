package swagger

import (
	"embed"
	"net/http"
)

//go:embed openapi.json
var specFS embed.FS

// ServeDoc serves the embedded OpenAPI document at /swagger/doc.json
func ServeDoc(w http.ResponseWriter, r *http.Request) {
	data, err := specFS.ReadFile("openapi.json")
	if err != nil {
		http.Error(w, "spec not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
