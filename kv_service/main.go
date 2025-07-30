package main

import (
	"encoding/json"
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
	log.Println("service line 24")
	store := make(map[string]string)
	for req := range reqCh {
		switch req.Method {
		case "set":
			log.Println("service, line 28")
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

var updatesToStoreChannel = make(chan KVRequest)

func setKey(c *gin.Context) {
	key := c.Param("key")

	var body struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil {
		http.Error(c.Writer, "invalid body", http.StatusBadRequest)
		return
	}

	setResponseChannel := make(chan KVResponse)
	updatesToStoreChannel <- KVRequest{Method: "set", Key: key, Value: body.Value, Resp: setResponseChannel} // pushes requests to channel being consumed by startKVStore function
	resp := <-setResponseChannel

	c.JSON(http.StatusCreated, gin.H{"key": key, "value": resp.Value})
}

func getKey(c *gin.Context) {
	log.Println("get key operation in progress")

	key := c.Param("key")
	getResponseChannel := make(chan KVResponse)
	updatesToStoreChannel <- KVRequest{Method: "get", Key: key, Resp: getResponseChannel} // pushes requests to channel being consumed by startKVStore function
	resp := <-getResponseChannel

	if !resp.Success {
		http.Error(c.Writer, "key not found", http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{"key": key, "value": resp.Value})
}

func deleteKey(c *gin.Context) {
	log.Println("delete key operation in progress")

	key := c.Param("key")
	deletionResponseChannel := make(chan KVResponse)
	updatesToStoreChannel <- KVRequest{Method: "delete", Key: key, Resp: deletionResponseChannel} // pushes requests to channel being consumed by startKVStore function
	resp := <-deletionResponseChannel

	c.JSON(http.StatusOK, gin.H{"whatwhat!": resp.Success})
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.PATCH("/kv/:key", setKey) // In this simple environment, setting a key is always a PUT because there's only one value that is always overwrititen; there's nothing to partially update. BUT, PUT makes me nervous and in a real production environment with larger objects, PATCH would be safer.
	r.GET("/kv/:key", getKey)
	r.DELETE("/kv/:key", deleteKey)
	r.POST("/kv/", func(c *gin.Context) { // skipping POST since they key is always known.
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method Not Allowed"})
	})
	return r
}

func main() {

	go startKVStore(updatesToStoreChannel) // using a goroutine to avoid race behavior writing/deleting the same key at the same time.

	r := setupRouter()

	log.Println("KV Service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
