package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	testCases := []struct {
		description        string
		input              []byte
		expectedStatusCode int
		method             string
		key                string
		expectedResponse   string
		needsStore         bool
	}{

		{
			description:        "success: GET key/value",
			input:              []byte(`{}`),
			expectedStatusCode: http.StatusOK,
			method:             http.MethodGet,
			key:                "robin",
			expectedResponse:   "{\"key\":\"robin\",\"value\":\"jasonTodd\"}",
			needsStore:         true,
		},
		{
			description:        "success: PATCH creates new key/value when key does not previously exist", //for maximum effect in a production environment, I would refactor so this test could include an initial GET to demonstrate the key does not exist before the operation. For today, observe the data seeded in the test includes no toast.
			input:              []byte(`{"value": "bruceWayne"}`),
			method:             http.MethodPatch,
			key:                "batman",
			expectedStatusCode: http.StatusCreated,
			expectedResponse:   "{\"key\":\"batman\",\"value\":\"bruceWayne\"}",
		},
		{
			description:        "success: PATCH key/value when exists",
			input:              []byte(`{"value": "dickGrayson"}`),
			method:             http.MethodPatch,
			key:                "robin",
			expectedStatusCode: http.StatusCreated,
			expectedResponse:   "{\"key\":\"robin\",\"value\":\"dickGrayson\"}",
			needsStore:         true,
		},
		{
			description:        "success: DELETE key/value",
			method:             http.MethodDelete,
			key:                "robin",
			expectedStatusCode: http.StatusOK,
			expectedResponse:   "{\"deletion successful for key\":\"robin\"}",
			needsStore:         true,
		},
		{
			description:        "error: GET, key does not exist",
			expectedStatusCode: http.StatusNotFound,
			method:             http.MethodGet,
			key:                "bigfoot",
			expectedResponse:   "{\"error\":\"key \\\"bigfoot\\\" not found\"}",
		},
		{
			description:        "error: DELETE, key does not exist", // could refactor to make DELETE on non existent resource a success: user wants resource not to exist; it does not ergo, what's the problem? what if user wanted to delete "katharine" which does exist, but supplied "katherine" which does not? I err on side of providing more information in an error.
			expectedStatusCode: http.StatusNotFound,
			method:             http.MethodDelete,
			key:                "nessie",
			expectedResponse:   "{\"error\":\"deletion unsuccessful for key: nessie\"}",
		},
		{
			description:        "error: PATCH request invalid",
			input:              []byte(`{"garbageData"}`),
			method:             http.MethodPatch,
			key:                "batman",
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"invalid body\"}",
		},
		{
			description:        "error: POST not supported",
			input:              []byte(`{"key": "superman","value": "clarkKent"}`),
			method:             http.MethodPost,
			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedResponse:   "{\"error\":\"Method Not Allowed\"}",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			updatesToStoreChannel := make(chan KVRequest)

			go startKVStore(updatesToStoreChannel)

			if tc.needsStore == true {
				seedResp := make(chan KVResponse)
				updatesToStoreChannel <- KVRequest{
					Method: "set",
					Key:    "robin",
					Value:  "jasonTodd",
					Resp:   seedResp,
				}
				<-seedResp
			}

			testRouter := setupRouter(updatesToStoreChannel)

			req, err := http.NewRequest(tc.method, "/kv/"+tc.key, bytes.NewBuffer(tc.input))
			assert.NoError(t, err)

			rw := httptest.NewRecorder()
			testRouter.ServeHTTP(rw, req)

			assert.Equal(t, tc.expectedStatusCode, rw.Code)
			assert.Equal(t, tc.expectedResponse, rw.Body.String())
		})
	}

}
