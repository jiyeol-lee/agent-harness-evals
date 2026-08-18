// Package httpclient calls the payments provider.
package httpclient

import (
	"fmt"
	"net/http"
	"time"
)

const (
	maxRetries   = 3
	retryBackoff = 200 * time.Millisecond
)

// Client charges orders through the payments provider.
type Client struct {
	http *http.Client
	base string
}

// New returns a client for the payments provider at base.
func New(base string) *Client {
	return &Client{http: &http.Client{Timeout: 5 * time.Second}, base: base}
}

// Charge charges the order with the given ID, retrying server errors.
func (c *Client) Charge(orderID string) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryBackoff)
		}

		resp, err := c.http.Post(fmt.Sprintf("%s/charge/%s", c.base, orderID), "application/json", nil)
		if err != nil {
			lastErr = err
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("charge %s: status %d", orderID, resp.StatusCode)
			continue
		}
		if resp.StatusCode >= 300 {
			return fmt.Errorf("charge %s: status %d", orderID, resp.StatusCode)
		}
		return nil
	}
	return lastErr
}
