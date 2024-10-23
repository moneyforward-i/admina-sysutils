package admina

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/moneyforward-i/admina-sysutils/internal/logger"
)

const (
	DefaultBaseURL = "https://api.itmc.i.moneyforward.com/api/v1"
)

type Client struct {
	BaseURL        string
	HTTPClient     *http.Client
	organizationID string
	apiKey         string
	Debug          bool
}

func NewClient(debug bool) *Client {
	baseURL := os.Getenv("ADMINA_BASE_URL")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	return &Client{
		BaseURL:        baseURL,
		HTTPClient:     &http.Client{},
		organizationID: os.Getenv("ADMINA_ORGANIZATION_ID"),
		apiKey:         os.Getenv("ADMINA_API_KEY"),
		Debug:          debug,
	}
}

func (c *Client) SetDebug(debug bool) {
	c.Debug = debug
}

func (c *Client) Get(path string, query map[string]string) (*http.Response, error) {
	url := fmt.Sprintf("%s/organizations/%s%s", c.BaseURL, c.organizationID, path)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for key, value := range query {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	logger.Debug.Printf("Sending GET request to %s", req.URL.String())
	logger.Debug.Printf("Headers: %v", req.Header)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		logger.Debug.Printf("Error sending request: %v", err)
		return nil, err
	}

	logger.Debug.Printf("Received response with status code %d", resp.StatusCode)

	return resp, nil
}

type Identity struct {
	ManagementType string `json:"managementType"`
	EmployeeStatus string `json:"employeeStatus"`
	Status         string `json:"status"`
}

type IdentityResponse struct {
	Data []Identity `json:"data"`
	Meta struct {
		TotalCount int    `json:"totalCount"`
		NextCursor string `json:"nextCursor"`
	} `json:"meta"`
}

func (c *Client) GetIdentities(cursor string) ([]Identity, string, int, error) {
	query := make(map[string]string)
	if cursor != "" {
		query["cursor"] = cursor
	}

	resp, err := c.Get("/identity", query)
	if err != nil {
		return nil, "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", 0, fmt.Errorf("API request error: %s", resp.Status)
	}

	var response struct {
		Data []Identity `json:"data"`
		Meta struct {
			TotalCount int    `json:"totalCount"`
			NextCursor string `json:"nextCursor"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, "", 0, err
	}

	return response.Data, response.Meta.NextCursor, response.Meta.TotalCount, nil
}

func (c *Client) IsDebug() bool {
	return c.Debug
}

func (c *Client) Validate() error {
	if c.organizationID == "" {
		return fmt.Errorf("ADMINA_ORGANIZATION_ID is not set")
	}
	if c.apiKey == "" {
		return fmt.Errorf("ADMINA_API_KEY is not set")
	}
	return nil
}
