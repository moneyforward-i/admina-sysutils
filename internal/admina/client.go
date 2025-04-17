package admina

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/moneyforward-i/admina-sysutils/internal/logger"
)

const (
	DefaultBaseURL = "https://api.itmc.i.moneyforward.com/api/v1"
	defaultTimeout = 30 * time.Second
)

// Client handles communication with the Admina API.
type Client struct {
	baseURL        string
	httpClient     *http.Client
	organizationID string
	apiKey         string
}

// NewClient creates a new Admina API client with default configuration.
func NewClient() *Client {
	baseURL := os.Getenv("ADMINA_BASE_URL")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	// Log proxy information if set
	proxyURLStr := os.Getenv("HTTPS_PROXY")
	if proxyURLStr == "" {
		proxyURLStr = os.Getenv("HTTP_PROXY")
	}

	if proxyURLStr != "" {
		parsedURL, err := url.Parse(proxyURLStr)
		if err != nil {
			logger.LogWarning("Failed to parse proxy URL from environment variable (%s): %v", proxyURLStr, err)
		} else {
			// Remove user info (credentials) before logging
			parsedURL.User = nil
			logger.LogInfo("Using proxy: %s", parsedURL.String())
		}
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   defaultTimeout,
			KeepAlive: defaultTimeout,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout:   defaultTimeout,
			Transport: transport,
		},
		organizationID: os.Getenv("ADMINA_ORGANIZATION_ID"),
		apiKey:         os.Getenv("ADMINA_API_KEY"),
	}
}

// APIError represents an error response from the Admina API.
type APIError struct {
	StatusCode    int
	Message       string
	Body          string
	RequestID     string
	Timestamp     time.Time
	OriginalError error
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error: status=%d, message=%s, body=%s, requestID=%s, timestamp=%s",
		e.StatusCode, e.Message, e.Body, e.RequestID, e.Timestamp.Format(time.RFC3339))
}

func (c *Client) debugLog(format string, args ...interface{}) {
	logger.LogDebug(format, args...)
}

func (c *Client) doRequest(ctx context.Context, method, path string, query map[string]string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s/organizations/%s%s", c.baseURL, c.organizationID, path)

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	if query != nil {
		q := req.URL.Query()
		for key, value := range query {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.debugLog("Request: %s %s", method, req.URL.String())
	if query != nil {
		c.debugLog("Query Params: %v", query)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.debugLog("Request Error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (c *Client) handleResponse(resp *http.Response) error {
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    resp.Status,
			Body:       string(body),
		}
	}
	return nil
}

// APIレスポンス関連の構造体
type Meta struct {
	StatusCode   int    `json:"statusCode"`
	ErrorCode    string `json:"errorCode,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	NextCursor   string `json:"nextCursor,omitempty"`
	TotalCount   int    `json:"totalCount,omitempty"`
	ItemsPerPage int    `json:"itemsPerPage,omitempty"`
	CurrentPage  int    `json:"currentPage,omitempty"`
}

type APIResponse[T any] struct {
	Meta  Meta `json:"meta"`
	Items T    `json:"items,omitempty"`
}

// Identity関連の構造体と関数
type Identity struct {
	ID              string   `json:"id"`
	OrganizationID  int      `json:"organizationId"`
	PeopleID        int      `json:"peopleId"`
	DisplayName     string   `json:"displayName"`
	ManagementType  string   `json:"managementType"`
	EmployeeType    string   `json:"employeeType"`
	EmployeeStatus  string   `json:"employeeStatus"`
	Email           string   `json:"primaryEmail"`
	SecondaryEmails []string `json:"secondaryEmails"`
	MergedPeople    []struct {
		ID           int    `json:"id"`
		DisplayName  string `json:"displayName"`
		PrimaryEmail string `json:"primaryEmail"`
		Username     string `json:"username"`
	} `json:"mergedPeople,omitempty"`
}

func (c *Client) GetIdentities(ctx context.Context, cursor string) ([]Identity, string, error) {
	query := map[string]string{}
	if cursor != "" {
		query["cursor"] = cursor
	}

	resp, err := c.doRequest(ctx, http.MethodGet, "/identity", query, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch identities: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound && cursor != "" {
		return []Identity{}, "", nil
	}

	if err := c.handleResponse(resp); err != nil {
		return nil, "", err
	}

	// レスポンスボディをバッファに読み込む
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	// RAWデータをログ出力
	c.debugLog("Raw Response Body: %s", string(bodyBytes))

	// 新しいReaderを作成してデコード
	var response APIResponse[[]Identity]
	if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&response); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Meta.ErrorCode != "" {
		return nil, "", fmt.Errorf("API error: %s - %s",
			response.Meta.ErrorCode,
			response.Meta.ErrorMessage)
	}

	c.debugLog("Retrieved %d identities", len(response.Items))
	return response.Items, response.Meta.NextCursor, nil
}

// Merge関連の構造体と関数
type MergeIdentity struct {
	FromPeopleID int `json:"fromPeopleId"`
	ToPeopleID   int `json:"toPeopleId"`
}

type MergeIdentityRequest struct {
	Merges []MergeIdentity `json:"merges"`
}

func (c *Client) MergeIdentities(ctx context.Context, fromPeopleID, toPeopleID int) (MergeIdentity, error) {
	payload := MergeIdentityRequest{
		Merges: []MergeIdentity{
			{
				FromPeopleID: fromPeopleID,
				ToPeopleID:   toPeopleID,
			},
		},
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/identity/merge", nil, payload)
	if err != nil {
		return MergeIdentity{}, fmt.Errorf("failed to merge identities: %w", err)
	}
	defer resp.Body.Close()

	if err := c.handleResponse(resp); err != nil {
		return MergeIdentity{}, err
	}

	// レスポンスボディをバッファに読み込む
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return MergeIdentity{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// RAWデータをログ出力
	c.debugLog("Raw Response Body: %s", string(bodyBytes))

	// APIからのレスポンス形式に合わせて構造体を定義
	var response APIResponse[[]Identity]

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		if len(bodyBytes) == 0 || string(bodyBytes) == "{}" || string(bodyBytes) == "[]" {
			c.debugLog("Received empty response, assuming merge success")
			return MergeIdentity{
				FromPeopleID: fromPeopleID,
				ToPeopleID:   toPeopleID,
			}, nil
		}
		return MergeIdentity{}, fmt.Errorf("failed to decode merge result: %w", err)
	}

	// レスポンスからマージされた情報を抽出
	if len(response.Items) == 0 {
		c.debugLog("Received empty items array, assuming merge success")
		return MergeIdentity{
			FromPeopleID: fromPeopleID,
			ToPeopleID:   toPeopleID,
		}, nil
	}

	// toPeopleID に該当するアイテムを探す
	for _, item := range response.Items {
		if item.PeopleID == toPeopleID {
			// マージされたアイテムが見つかった
			// mergedPeople 配列から fromPeopleID に該当する情報を確認
			for _, mergedPerson := range item.MergedPeople {
				if mergedPerson.ID == fromPeopleID {
					return MergeIdentity{
						FromPeopleID: fromPeopleID,
						ToPeopleID:   toPeopleID,
					}, nil
				}
			}
		}
	}

	// 該当するマージ情報が見つからない場合（通常はここに来ないはず）
	c.debugLog("Could not find merged identity in response, assuming success based on request")
	return MergeIdentity{
		FromPeopleID: fromPeopleID,
		ToPeopleID:   toPeopleID,
	}, nil
}

// Organization関連の構造体と関数
type Organization struct {
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	UniqueName      string   `json:"uniqueName"`
	Status          string   `json:"status"`
	SystemLanguage  string   `json:"systemLanguage"`
	Location        string   `json:"location"`
	TimeZone        string   `json:"timeZone"`
	Domains         []string `json:"domains"`
	ForwardingEmail string   `json:"forwardingEmail"`
	TrialCount      int      `json:"trialCount"`
}

func (c *Client) GetOrganization(ctx context.Context) (*Organization, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	defer resp.Body.Close()

	if err := c.handleResponse(resp); err != nil {
		return nil, err
	}

	var org Organization
	if err := json.NewDecoder(resp.Body).Decode(&org); err != nil {
		return nil, fmt.Errorf("failed to decode organization: %w", err)
	}

	return &org, nil
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

// CreateIdentityRequest represents the request body for creating an identity
type CreateIdentityRequest struct {
	PrimaryEmail   string `json:"primaryEmail"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	DisplayName    string `json:"displayName"`
	EmployeeType   string `json:"employeeType"`
	EmployeeStatus string `json:"employeeStatus"`
}

// CreateIdentity creates a new identity
func (c *Client) CreateIdentity(ctx context.Context, req *CreateIdentityRequest) (*Identity, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/identity", nil, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity: %w", err)
	}
	defer resp.Body.Close()

	if err := c.handleResponse(resp); err != nil {
		return nil, err
	}

	var identity Identity
	if err := json.NewDecoder(resp.Body).Decode(&identity); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &identity, nil
}

// DeleteIdentity deletes an identity by ID
func (c *Client) DeleteIdentity(ctx context.Context, identityID string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/identity/%s", identityID), nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete identity: %w", err)
	}
	defer resp.Body.Close()

	if err := c.handleResponse(resp); err != nil {
		return err
	}

	return nil
}
