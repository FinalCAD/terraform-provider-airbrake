package airbrake

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type Client struct {
	BaseUrl string
	Token   string `json:"-"`
}

func NewClient(baseUrl string, email string, password string, apiKey string) (*Client, error) {
	client := &Client{BaseUrl: baseUrl, Token: ""}
	if apiKey != "" {
		client.Token = apiKey
		statusCode, _, err := client.Get("users/current", nil)
		if err != nil {
			return nil, err
		}
		if statusCode >= 300 {
			return nil, errors.New("users/current wrong api key")
		}
	} else {
		params := map[string]string{
			"email":    email,
			"password": password,
		}
		statusCode, responseBody, err := client.Get("sessions", params)
		if err != nil {
			return nil, err
		}
		if statusCode >= 300 {
			return nil, errors.New("sessions : wrong login/password")
		}

		var response map[string]interface{}
		err = json.Unmarshal(responseBody, &response)
		if err != nil {
			return nil, err
		}

		token, ok := response["token"].(string)
		if !ok {
			return nil, errors.New("token not found")
		}
		client.Token = token
	}
	return client, nil
}

func (c *Client) Get(domain string, params map[string]string) (int, []byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, c.BaseUrl+domain, nil)
	if err != nil {
		return 500, nil, err
	}

	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	if c.Token != "" {
		q.Add("key", c.Token)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return 404, nil, err
	}

	responseBody, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return 500, nil, err
	}

	return resp.StatusCode, responseBody, nil
}

func (c *Client) Post(domain string, params map[string]string) (int, []byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, c.BaseUrl+domain, nil)
	if err != nil {
		return 500, nil, err
	}

	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	if c.Token != "" {
		q.Add("key", c.Token)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return 404, nil, err
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 500, nil, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, responseBody, nil
}

func (c *Client) Put(domain string, id string, values map[string]string) (int, error) {
	client := &http.Client{}

	data, err := json.Marshal(values)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest(http.MethodPut, c.BaseUrl+domain+"/"+id, bytes.NewBuffer(data))
	if err != nil {
		return 500, err
	}

	q := req.URL.Query()
	if c.Token != "" {
		q.Add("key", c.Token)
	}

	req.Header.Set("Content-Type", "application/json")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return 500, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func (c *Client) Delete(domain string, id string) (int, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, c.BaseUrl+domain+"/"+id, nil)
	if err != nil {
		return 500, err
	}

	q := req.URL.Query()
	if c.Token != "" {
		q.Add("key", c.Token)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return 500, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
