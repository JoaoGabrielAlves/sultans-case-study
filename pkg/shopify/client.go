package shopify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ShopifyURL     string
	AccessToken    string
	UserEmail      string
	StorePassword  string
}

type Client struct {
	httpClient *http.Client
	endpoint   string
	config     *Config
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message   string `json:"message"`
		Locations []struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		} `json:"locations"`
		Path []interface{} `json:"path"`
	} `json:"errors"`
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

func NewClient() (*Client, error) {
	godotenv.Load()
	
	config := &Config{
		ShopifyURL:     os.Getenv("SHOPIFY_URL"),
		AccessToken:    os.Getenv("SHOPIFY_ACCESS_TOKEN"),
		UserEmail:      os.Getenv("SHOPIFY_USER_EMAIL"),
		StorePassword:  os.Getenv("SHOPIFY_STORE_PASSWORD"),
	}

	if config.ShopifyURL == "" || config.AccessToken == "" {
		return nil, fmt.Errorf("missing required environment variables: SHOPIFY_URL and SHOPIFY_ACCESS_TOKEN are required")
	}

	endpoint := fmt.Sprintf("%s/admin/api/2025-01/graphql.json", config.ShopifyURL)
	
	return &Client{
		httpClient: &http.Client{},
		endpoint:   endpoint,
		config:     config,
	}, nil
}

func (c *Client) GraphQL(ctx context.Context, query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	request := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Access-Token", c.config.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Errors) > 0 {
		return &response, fmt.Errorf("GraphQL errors: %v", response.Errors)
	}

	return &response, nil
}

func (c *Client) GetConfig() *Config {
	return c.config
} 