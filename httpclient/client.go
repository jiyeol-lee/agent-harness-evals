// Package httpclient calls the payments provider.
package httpclient

import (
	"fmt"
	"net/http"
	"time"
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

// Charge charges the order with the given ID.
func (c *Client) Charge(orderID string) error {
	resp, err := c.http.Post(fmt.Sprintf("%s/charge/%s", c.base, orderID), "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("charge %s: status %d", orderID, resp.StatusCode)
	}
	return nil
}
