# Key-Value Store (Go)

This is a basic in-memory key-value store written in Go. It supports simple HTTP operations to get, set, and delete values by key.
The repo includes unit tests and a simple HTTP test client that runs in a separate container and issues predefined test requests to the server.

Some author notes: there are code comments in there beyond what I would commit to production , because in the context of a technical evaluation, I selfishly want to explain my thought process with some of the decisions I made.  Additionally, the server uses Gin, which is my preferred Go server setup; I switched to Mux for the client because I thought it would be a little more lightweight to implement; I wanted to demonstrate flexibility; and it was fun for me. In a production environment teammates I would prioritize choosing one over the other for consistency and ease of collaboration.  

Thank you for your consideration. I hope I will have the opportunity to speak more to you about this project and the thought-process behind it in the near future. 

## Requirements

- Go 1.18 or higher
- Docker (for running the full stack locally)

## Setup (for local development)

Install dependencies:

```bash
go mod tidy
```

## Running the App with Docker

To start both the key-value store and the test client:

```bash
docker-compose up --build
```

- The service will run on `localhost:8080`
- The test client will run on `localhost:8081`



## HTTP Endpoints

- `GET /kv/{key}` – Get the value associated with `{key}`
- `PATCH /kv/{key}` – Create or update the value for `{key}`
- `DELETE /kv/{key}` – Remove `{key}` and its value


## Using the Test Client

With both services running via Docker, try the following test cases:

```bash
curl -i -X GET -H "Content-Type: application/json" localhost:8081/test_overwrite
curl -i -X GET -H "Content-Type: application/json" localhost:8081/test_deletion
``` 

You can also call the supported HTTP methods directly:

```bash
# Create or update a key
curl -i -X PATCH -H "Content-Type: application/json" -d '{"value": "bruce wayne"}' localhost:8081/kv/batman

# Get the value for a specified key
curl -i -X GET -H "Content-Type: application/json" localhost:8081/kv/batman

# Delete the key
curl -i -X DELETE -H "Content-Type: application/json" localhost:8081/kv/batman
```

## How to Run Tests
Unit tests are located in `service_test.go` and cover both happy-path and error cases."


To run the tests:

```bash
go test
```

The tests cover:

- Valid GET, PATCH, and DELETE operations
- Error handling (e.g. unsupported methods, invalid input)


## Notes

- Uses Go channels for safe concurrent updates
- Test coverage includes both success and failure paths
- Built with the standard library, plus:
  - [`gin-gonic/gin`](https://github.com/gin-gonic/gin) – HTTP router
  - [`gorilla/mux`](https://github.com/gorilla/mux) – a different HTTP router
  - [`stretchr/testify`](https://github.com/stretchr/testify) – Testing utilities