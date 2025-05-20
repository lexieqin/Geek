package utils

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

// GetHTTP executes a GET HTTP request to the specified URL and returns the response body.
func GetHTTP(url string) (string, error) {
	// Create HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func PostHTTP(url string, body []byte) (string, error) {
	// Create HTTP POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}

// DeleteHTTP executes a DELETE HTTP request to the specified URL and returns the response body.
func DeleteHTTP(url string) (string, error) {
	// Create HTTP DELETE request
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}
