package identity

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	"github.com/moneyforward-i/admina-sysutils/internal/logger"
)

type mockTransport struct {
	response *http.Response
	err      error
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func setupMockClient(response map[string]interface{}) *http.Client {
	respBody, _ := json.Marshal(response)
	mockResp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBuffer(respBody)),
		Header:     make(http.Header),
	}
	mockResp.Header.Set("Content-Type", "application/json")

	return &http.Client{
		Transport: &mockTransport{
			response: mockResp,
		},
	}
}

func mockIdentityResponse() map[string]interface{} {
	return map[string]interface{}{
		"meta": map[string]interface{}{
			"nextCursor": "",
			"totalCount": 10,
		},
		"data": []map[string]interface{}{
			{
				"employeeStatus": "active",
				"managementType": "managed",
				"employeeType":   "full_time_employee",
			},
			{
				"employeeStatus": "active",
				"managementType": "external",
				"employeeType":   "part_time_employee",
			},
			{
				"employeeStatus": "draft",
				"managementType": "system",
				"employeeType":   "temporary_employee",
			},
			{
				"employeeStatus": "preactive",
				"managementType": "unknown",
				"employeeType":   "contract_employee",
			},
			{
				"employeeStatus": "retired",
				"managementType": "unregistered",
				"employeeType":   "secondment_employee",
			},
			{
				"employeeStatus": "untracked",
				"managementType": "managed",
				"employeeType":   "fixed_time_employee",
			},
			{
				"employeeStatus": "archived",
				"managementType": "external",
				"employeeType":   "collaborator",
			},
			{
				"employeeStatus": "active",
				"managementType": "system",
				"employeeType":   "board_member",
			},
			{
				"employeeStatus": "on_leave",
				"managementType": "unknown",
				"employeeType":   "group_address",
			},
			{
				"employeeStatus": "draft",
				"managementType": "unregistered",
				"employeeType":   "shared_address",
			},
		},
	}
}

func init() {
	debugMode, _ := strconv.ParseBool(os.Getenv("ADMINA_TEST_DEBUG"))
	logger.Init(debugMode)
}

func getTestDebugMode() bool {
	debugMode, _ := strconv.ParseBool(os.Getenv("ADMINA_TEST_DEBUG"))
	return debugMode
}

func TestGetIdentityMatrix(t *testing.T) {
	mockClient := &MockIdentityClient{
		identities: []Identity{
			{ManagementType: "Employee", EmployeeStatus: "active"},
			{ManagementType: "Employee", EmployeeStatus: "active"},
			{ManagementType: "Contractor", EmployeeStatus: "active"},
			{ManagementType: "Employee", EmployeeStatus: "on_leave"},
		},
		cursor: "",
	}

	expectedMatrix := &Matrix{
		ManagementTypes: []string{"Employee", "Contractor"},
		Statuses:        []string{"active", "on_leave"},
		Matrix: [][]int{
			{2, 1},
			{1, 0},
		},
	}

	result, err := GetIdentityMatrix(mockClient)
	if err != nil {
		t.Fatalf("GetIdentityMatrix returned an error: %v", err)
	}

	if !reflect.DeepEqual(result, expectedMatrix) {
		t.Errorf("GetIdentityMatrix returned unexpected result. Got %+v, want %+v", result, expectedMatrix)
	}
}

func TestPrintIdentityMatrix(t *testing.T) {
	organizationID := os.Getenv("ADMINA_ORGANIZATION_ID")
	apiKey := os.Getenv("ADMINA_API_KEY")
	if organizationID == "" || apiKey == "" {
		t.Fatal("ADMINA_ORGANIZATION_ID or ADMINA_API_KEY is not set")
	}

	mockClient := setupMockClient(mockIdentityResponse())
	adminaClient := admina.NewClient(getTestDebugMode())
	adminaClient.HTTPClient = mockClient

	// Use identityClientAdapter
	client := &identityClientAdapter{client: adminaClient}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := PrintIdentityMatrix(client, "pretty")
	if err != nil {
		t.Errorf("PrintIdentityMatrix() failed with error: %v", err)
	}

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check for expected output parts
	expectedOutputs := []string{
		"Identity Matrix:",
		"managed",
		"external",
		"system",
		"unknown",
		"unregistered",
		"draft",
		"preactive",
		"active",
		"on_leave",
		"retired",
		"untracked",
		"archived",
		"Total",
		"10", // Total Identity count
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output '%s' not found", expected)
		}
	}
}

// Add identityClientAdapter definition to the test file
type identityClientAdapter struct {
	client *admina.Client
}

func (a *identityClientAdapter) GetIdentities(cursor string) ([]Identity, string, int, error) {
	adminaIdentities, nextCursor, totalCount, err := a.client.GetIdentities(cursor)
	if err != nil {
		return nil, "", 0, err
	}

	identities := make([]Identity, len(adminaIdentities))
	for i, ai := range adminaIdentities {
		identities[i] = Identity{
			ManagementType: ai.ManagementType,
			EmployeeStatus: ai.EmployeeStatus,
			Status:         ai.Status,
		}
	}

	return identities, nextCursor, totalCount, nil
}

type MockIdentityClient struct {
	identities []Identity
	cursor     string
}

func (m *MockIdentityClient) GetIdentities(cursor string) ([]Identity, string, int, error) {
	return m.identities, m.cursor, len(m.identities), nil
}
