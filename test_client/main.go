package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type InboundRequest struct {
	Value string `json:"value"`
}

func callKV(method, path string, body any) (*http.Response, error) {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, "http://kv_service:8080"+path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

func handler(w http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)
	key := vars["key"]

	var inboundRequest InboundRequest
	var err error
	var resp *http.Response
	var outboundReq map[string]string

	if req.Method == "PATCH" {
		err = json.NewDecoder(req.Body).Decode(&inboundRequest)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "could not decode JSON request"})
			return
		}
		outboundReq = map[string]string{"value": inboundRequest.Value}
	}

	resp, err = callKV(req.Method, "/kv/"+key, outboundReq)

	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{"error": "call failed"})
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		var successfulResponse struct {
			Success bool
			Value   string
		}

		json.NewDecoder(resp.Body).Decode(&successfulResponse)
		json.NewEncoder(w).Encode(map[string]string{"key": key, "value": successfulResponse.Value})
		return
	}

	var errResp struct {
		Error string `json:"error"`
	}

	json.NewDecoder(resp.Body).Decode(&errResp)
	json.NewEncoder(w).Encode(map[string]string{"error": errResp.Error})

}

func testDeletionHandler(w http.ResponseWriter, req *http.Request) {
	if _, err := callKV("PATCH", "/kv/batman", `{"value": "bruce wayne"}`); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to set key", "details": err.Error()})
		return
	}

	if _, err := callKV("DELETE", "/kv/batman", nil); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete key", "details": err.Error()})
		return
	}

	if _, err := callKV("GET", "/kv/batman", nil); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Key still exists after deletion"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "test overwrite passed"})

}

func testOverWriteHandler(w http.ResponseWriter, req *http.Request) {

	if _, err := callKV("PATCH", "/kv/robin", `{"value": "dick grayson"}`); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to set key", "details": err.Error()})
		return
	}
	if _, err := callKV("PATCH", "/kv/robin", `{"value": "jason todd"}`); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to set key", "details": err.Error()})
		return
	}

	resp, err := callKV("GET", "/kv/robin", nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Key still exists after deletion"})
		return
	}

	var successfulResponse struct {
		Success bool
		Value   string
	}

	json.NewDecoder(resp.Body).Decode(&successfulResponse)

	if successfulResponse.Value != "jason todd" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Overwrite test failed", "expected": "jason todd", "got": successfulResponse.Value})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "test overwrite passed"})

}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/kv/{key}", handler).Methods("PATCH", "DELETE", "GET")
	r.HandleFunc("/test_deletion", testDeletionHandler).Methods("GET")
	r.HandleFunc("/test_overwrite", testOverWriteHandler).Methods("GET")

	log.Println("Test Client listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}
