package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type KVRequest struct { // FUTURE: wouldn't it be great if this were a protobuf defined in a shared repository so users can see how their requests should be formatted?
	Method string
	Key    string
	Value  string
	Resp   chan KVResponse
}

type KVResponse struct {
	Success bool
	Value   string
}

func startKVStore(reqCh <-chan KVRequest) {
	store := make(map[string]string)
	for req := range reqCh {
		switch req.Method {
		case "set":
			store[req.Key] = req.Value
			req.Resp <- KVResponse{Success: true, Value: req.Value}
		case "get":
			val, ok := store[req.Key]
			req.Resp <- KVResponse{Success: ok, Value: val}
		case "delete":
			_, ok := store[req.Key]
			delete(store, req.Key)
			req.Resp <- KVResponse{Success: ok}
		}
	}
}

type KVHandler struct {
	storeChannel chan KVRequest
}

func (h *KVHandler) setKey(c *gin.Context) {
	key := c.Param("key")
	log.Printf(`set operation in progress for key: %s`, key)

	var requestBody struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&requestBody); err != nil || requestBody.Value == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	setResponseChannel := make(chan KVResponse)
	h.storeChannel <- KVRequest{Method: "set", Key: key, Value: requestBody.Value, Resp: setResponseChannel} // pushes requests to channel being consumed by startKVStore function
	resp := <-setResponseChannel

	c.JSON(http.StatusCreated, gin.H{"key": key, "value": resp.Value})
}

func (h *KVHandler) getKey(c *gin.Context) {
	key := c.Param("key")
	log.Printf(`get operation in progress for key: %s`, key)

	getResponseChannel := make(chan KVResponse)
	h.storeChannel <- KVRequest{Method: "get", Key: key, Resp: getResponseChannel}
	resp := <-getResponseChannel

	if !resp.Success {
		errMessage := fmt.Sprintf(`key "%s" not found`, key)
		c.JSON(http.StatusNotFound, gin.H{"error": errMessage})
		return
	}

	c.JSON(http.StatusOK, gin.H{"key": key, "value": resp.Value})
}

func (h *KVHandler) deleteKey(c *gin.Context) {
	key := c.Param("key")
	log.Printf(`delete operation in progress for key: %s`, key)

	deletionResponseChannel := make(chan KVResponse)
	h.storeChannel <- KVRequest{Method: "delete", Key: key, Resp: deletionResponseChannel}
	resp := <-deletionResponseChannel

	if resp.Success == false {
		errMessage := fmt.Sprintf(`deletion unsuccessful for key: %s`, key)
		c.JSON(http.StatusNotFound, gin.H{"error": errMessage})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deletion successful for key": key})
}

func setupRouter(storeChannel chan KVRequest) *gin.Engine {
	r := gin.Default()
	handler := &KVHandler{storeChannel: storeChannel}

	r.PATCH("/kv/:key", handler.setKey) // In this simple environment, setting a key is always a PUT because there's only one value that is always overwritten; there's nothing to partially update. BUT, PUT makes me nervous and in a real production environment with larger objects, PATCH would be safer.
	r.GET("/kv/:key", handler.getKey)
	r.DELETE("/kv/:key", handler.deleteKey)
	r.POST("/kv/", func(c *gin.Context) { // skipping POST since the key is always known.
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method Not Allowed"})
	})
	return r
}

var updatesToStoreChannel = make(chan KVRequest)

func main() {

	go startKVStore(updatesToStoreChannel) // using a goroutine to avoid race behavior writing/deleting the same key at the same time.

	r := setupRouter(updatesToStoreChannel)

	log.Println("KV Service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
