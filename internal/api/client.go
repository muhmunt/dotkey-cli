package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{},
	}
}

// do executes a request and returns the response body bytes.
func (c *Client) do(method, path string, body interface{}) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot reach EnvX API at %s\nCheck your connection or run `dotkey login` to reconfigure", c.baseURL)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, resp.StatusCode, errors.New("not logged in. Run `dotkey login` first")
	}

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(data, &apiErr) == nil && apiErr.Error != "" {
			return nil, resp.StatusCode, errors.New(apiErr.Error)
		}
		return nil, resp.StatusCode, fmt.Errorf("API error (status %d)", resp.StatusCode)
	}

	return data, resp.StatusCode, nil
}

func decode[T any](data []byte) (T, error) {
	var v T
	return v, json.Unmarshal(data, &v)
}

// ── Auth ─────────────────────────────────────────────────────────────────────

func (c *Client) Register(name, email, password string) (string, *User, error) {
	data, _, err := c.do("POST", "/api/v1/auth/register", map[string]string{
		"name": name, "email": email, "password": password,
	})
	if err != nil {
		return "", nil, err
	}
	var res struct {
		Token string `json:"token"`
		User  User   `json:"user"`
	}
	err = json.Unmarshal(data, &res)
	return res.Token, &res.User, err
}

func (c *Client) Login(email, password string) (string, error) {
	data, _, err := c.do("POST", "/api/v1/auth/login", map[string]string{
		"email": email, "password": password,
	})
	if err != nil {
		return "", err
	}
	var res struct {
		Token string `json:"token"`
	}
	return res.Token, json.Unmarshal(data, &res)
}

func (c *Client) Me() (*User, error) {
	data, _, err := c.do("GET", "/api/v1/auth/me", nil)
	if err != nil {
		return nil, err
	}
	return decode[*User](data)
}

func (c *Client) DeviceCode() (*DeviceCodeResult, error) {
	data, _, err := c.do("POST", "/api/v1/auth/device", nil)
	if err != nil {
		return nil, err
	}
	return decode[*DeviceCodeResult](data)
}

func (c *Client) DevicePoll(deviceCode string) (*DevicePollResult, error) {
	data, status, err := c.do("GET", "/api/v1/auth/device/poll?device_code="+deviceCode, nil)
	if status == http.StatusAccepted {
		return &DevicePollResult{Status: "pending"}, nil
	}
	if err != nil {
		return nil, err
	}
	return decode[*DevicePollResult](data)
}

// ── Projects ─────────────────────────────────────────────────────────────────

func (c *Client) ListProjects() ([]Project, error) {
	data, _, err := c.do("GET", "/api/v1/projects", nil)
	if err != nil {
		return nil, err
	}
	return decode[[]Project](data)
}

func (c *Client) CreateProject(name, description string) (*Project, error) {
	data, _, err := c.do("POST", "/api/v1/projects", map[string]string{
		"name": name, "description": description,
	})
	if err != nil {
		return nil, err
	}
	return decode[*Project](data)
}

func (c *Client) GetProject(id string) (*Project, error) {
	data, _, err := c.do("GET", "/api/v1/projects/"+id, nil)
	if err != nil {
		return nil, err
	}
	return decode[*Project](data)
}

func (c *Client) ListMembers(projectID string) ([]ProjectMember, error) {
	data, _, err := c.do("GET", "/api/v1/projects/"+projectID+"/members", nil)
	if err != nil {
		return nil, err
	}
	return decode[[]ProjectMember](data)
}

// ── Environments ─────────────────────────────────────────────────────────────

func (c *Client) ListEnvironments(projectID string) ([]Environment, error) {
	data, _, err := c.do("GET", "/api/v1/projects/"+projectID+"/environments", nil)
	if err != nil {
		return nil, err
	}
	return decode[[]Environment](data)
}

func (c *Client) CreateEnvironment(projectID, name string) (*Environment, error) {
	data, _, err := c.do("POST", "/api/v1/projects/"+projectID+"/environments", map[string]string{
		"name": name,
	})
	if err != nil {
		return nil, err
	}
	return decode[*Environment](data)
}

func (c *Client) DeleteEnvironment(projectID, envID string) error {
	_, _, err := c.do("DELETE", "/api/v1/projects/"+projectID+"/environments/"+envID, nil)
	return err
}

// ── Variables ─────────────────────────────────────────────────────────────────

func (c *Client) ListVariables(projectID, envID string) ([]Variable, error) {
	data, _, err := c.do("GET", "/api/v1/projects/"+projectID+"/environments/"+envID+"/variables", nil)
	if err != nil {
		return nil, err
	}
	return decode[[]Variable](data)
}

func (c *Client) CreateVariable(projectID, envID, key, value string) (*Variable, error) {
	data, _, err := c.do("POST", "/api/v1/projects/"+projectID+"/environments/"+envID+"/variables",
		map[string]string{"key": key, "value": value})
	if err != nil {
		return nil, err
	}
	return decode[*Variable](data)
}

func (c *Client) UpdateVariable(projectID, envID, varID, value string) (*Variable, error) {
	data, _, err := c.do("PUT", "/api/v1/projects/"+projectID+"/environments/"+envID+"/variables/"+varID,
		map[string]string{"value": value})
	if err != nil {
		return nil, err
	}
	return decode[*Variable](data)
}

func (c *Client) DeleteVariable(projectID, envID, varID string) error {
	_, _, err := c.do("DELETE", "/api/v1/projects/"+projectID+"/environments/"+envID+"/variables/"+varID, nil)
	return err
}

// Export returns raw KEY=VALUE content (decrypted).
func (c *Client) Export(projectID, envID string) (string, error) {
	req, _ := http.NewRequest("GET", c.baseURL+"/api/v1/projects/"+projectID+"/environments/"+envID+"/export", nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return "", errors.New("not logged in. Run `dotkey login` first")
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("failed to export variables (status %d)", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	return string(data), err
}

// Import sends KEY=VALUE content to the server (upserts all variables).
func (c *Client) Import(projectID, envID, content string) (*ImportResult, error) {
	data, _, err := c.do("POST", "/api/v1/projects/"+projectID+"/environments/"+envID+"/import",
		map[string]string{"content": content})
	if err != nil {
		return nil, err
	}
	return decode[*ImportResult](data)
}

// ── Diff ─────────────────────────────────────────────────────────────────────

func (c *Client) Diff(projectID, fromEnvID, toEnvID string) ([]DiffEntry, error) {
	data, _, err := c.do("GET", "/api/v1/projects/"+projectID+"/diff?from="+fromEnvID+"&to="+toEnvID, nil)
	if err != nil {
		return nil, err
	}
	return decode[[]DiffEntry](data)
}

// ── History & Rollback ───────────────────────────────────────────────────────

func (c *Client) History(projectID, envID string) ([]VariableVersion, error) {
	data, _, err := c.do("GET", "/api/v1/projects/"+projectID+"/environments/"+envID+"/history", nil)
	if err != nil {
		return nil, err
	}
	return decode[[]VariableVersion](data)
}

func (c *Client) Rollback(projectID, envID, versionID string) error {
	_, _, err := c.do("POST", "/api/v1/projects/"+projectID+"/environments/"+envID+"/rollback/"+versionID, nil)
	return err
}
