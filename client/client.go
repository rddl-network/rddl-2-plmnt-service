package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/rddl-network/rddl-2-plmnt-service/service"
)

type IR2PClient interface {
	GetReceiveAddress(ctx context.Context, plmntAddress string) (res service.ReceiveAddressResponse, err error)
}

type R2PClient struct {
	baseURL string
	client  *http.Client
}

func NewR2PClient(baseURL string, client *http.Client) *R2PClient {
	if client == nil {
		client = &http.Client{}
	}
	return &R2PClient{
		baseURL: baseURL,
		client:  client,
	}
}

func (r2pc *R2PClient) GetReceiveAddress(ctx context.Context, plmntAddress string) (res service.ReceiveAddressResponse, err error) {
	err = r2pc.doRequest(ctx, http.MethodGet, r2pc.baseURL+"/receiveaddress/"+plmntAddress, nil, &res)
	return
}

func (r2pc *R2PClient) doRequest(ctx context.Context, method, url string, body interface{}, response interface{}) (err error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := r2pc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return &httpError{StatusCode: resp.StatusCode}
	}

	if response != nil {
		return json.NewDecoder(resp.Body).Decode(response)
	}

	return
}

type httpError struct {
	StatusCode int
}

func (e *httpError) Error() string {
	return http.StatusText(e.StatusCode)
}
