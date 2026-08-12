// Package appium implements the Appium 2 W3C WebDriver HTTP client.
// UiAutomator2 is selected via appium:automationName capability.
package appium

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	HTTP       *http.Client
	SessionID  string
	Serial     string
}

type SessionOptions struct {
	ServerURL       string
	Serial          string
	AutomationName  string
	AppPackage      string
	AppActivity     string
	NoReset         bool
	AutoGrantPerms  bool
	NewCommandTimeout int
}

func NewClient(opts SessionOptions) *Client {
	url := strings.TrimRight(opts.ServerURL, "/")
	if url == "" {
		url = "http://127.0.0.1:4723"
	}
	return &Client{
		BaseURL: url,
		Serial:  opts.Serial,
		HTTP:    &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *Client) CreateSession(opts SessionOptions) error {
	auto := opts.AutomationName
	if auto == "" {
		auto = "UiAutomator2"
	}
	caps := map[string]interface{}{
		"platformName":              "Android",
		"appium:automationName":     auto,
		"appium:udid":               opts.Serial,
		"appium:noReset":            opts.NoReset,
		"appium:autoGrantPermissions": opts.AutoGrantPerms,
	}
	if opts.AppPackage != "" {
		caps["appium:appPackage"] = opts.AppPackage
	}
	if opts.AppActivity != "" {
		caps["appium:appActivity"] = opts.AppActivity
	}
	if opts.NewCommandTimeout > 0 {
		caps["appium:newCommandTimeout"] = opts.NewCommandTimeout
	}
	body := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"alwaysMatch": caps,
		},
	}
	var resp struct {
		Value struct {
			SessionID string `json:"sessionId"`
		} `json:"value"`
	}
	if err := c.postJSON("/session", body, &resp); err != nil {
		return fmt.Errorf("create appium session: %w", err)
	}
	c.SessionID = resp.Value.SessionID
	if c.SessionID == "" {
		return fmt.Errorf("appium returned empty session id")
	}
	return nil
}

func (c *Client) DeleteSession() error {
	if c.SessionID == "" {
		return nil
	}
	return c.delete(fmt.Sprintf("/session/%s", c.SessionID))
}

func (c *Client) FindByText(text string) (string, error) {
	// UiAutomator2 backend — primary selector strategy
	selector := fmt.Sprintf(`new UiSelector().text("%s")`, escapeSelector(text))
	return c.findElement("android uiautomator", selector)
}

func (c *Client) FindByTextContains(text string) (string, error) {
	selector := fmt.Sprintf(`new UiSelector().textContains("%s")`, escapeSelector(text))
	return c.findElement("android uiautomator", selector)
}

func (c *Client) FindByTexts(texts []string) (string, error) {
	var lastErr error
	for _, t := range texts {
		if t == "" {
			continue
		}
		id, err := c.FindByText(t)
		if err == nil {
			return id, nil
		}
		id, err = c.FindByTextContains(t)
		if err == nil {
			return id, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("element not found: %v", texts)
}

func (c *Client) findElement(using, value string) (string, error) {
	var resp struct {
		Value struct {
			ElementID string `json:"ELEMENT"`
		} `json:"value"`
	}
	// Appium 2 may return element id in "element-xxx" format in value directly as string
	var raw map[string]interface{}
	path := fmt.Sprintf("/session/%s/element", c.SessionID)
	if err := c.postJSON(path, map[string]string{"using": using, "value": value}, &raw); err != nil {
		return "", err
	}
	if v, ok := raw["value"].(map[string]interface{}); ok {
		for k, id := range v {
			if strings.Contains(k, "element") && id != nil {
				return fmt.Sprint(id), nil
			}
		}
	}
	if s, ok := raw["value"].(string); ok {
		return s, nil
	}
	_ = resp
	return "", fmt.Errorf("empty element id for %q", value)
}

func (c *Client) Click(elementID string) error {
	path := fmt.Sprintf("/session/%s/element/%s/click", c.SessionID, elementID)
	return c.postJSON(path, map[string]interface{}{}, &struct{}{})
}

func (c *Client) SendKeys(elementID, text string) error {
	path := fmt.Sprintf("/session/%s/element/%s/value", c.SessionID, elementID)
	return c.postJSON(path, map[string]interface{}{
		"text": text,
		"value": strings.Split(text, ""),
	}, &struct{}{})
}

func (c *Client) Tap(x, y int) error {
	path := fmt.Sprintf("/session/%s/actions", c.SessionID)
	body := map[string]interface{}{
		"actions": []map[string]interface{}{
			{
				"type": "pointer",
				"id":   "finger1",
				"parameters": map[string]string{"pointerType": "touch"},
				"actions": []map[string]interface{}{
					{"type": "pointerMove", "duration": 0, "x": x, "y": y},
					{"type": "pointerDown", "button": 0},
					{"type": "pause", "duration": 50},
					{"type": "pointerUp", "button": 0},
				},
			},
		},
	}
	return c.postJSON(path, body, &struct{}{})
}

func (c *Client) Swipe(x1, y1, x2, y2, durationMs int) error {
	path := fmt.Sprintf("/session/%s/actions", c.SessionID)
	body := map[string]interface{}{
		"actions": []map[string]interface{}{
			{
				"type": "pointer",
				"id":   "finger1",
				"parameters": map[string]string{"pointerType": "touch"},
				"actions": []map[string]interface{}{
					{"type": "pointerMove", "duration": 0, "x": x1, "y": y1},
					{"type": "pointerDown", "button": 0},
					{"type": "pointerMove", "duration": durationMs, "x": x2, "y": y2},
					{"type": "pointerUp", "button": 0},
				},
			},
		},
	}
	return c.postJSON(path, body, &struct{}{})
}

func (c *Client) PageSource() (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	path := fmt.Sprintf("/session/%s/source", c.SessionID)
	if err := c.getJSON(path, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

func (c *Client) postJSON(path string, body interface{}, out interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.HTTP.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out)
}

func (c *Client) getJSON(path string, out interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.HTTP.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *Client) delete(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	return c.do(req, &struct{}{})
}

func (c *Client) do(req *http.Request, out interface{}) error {
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("appium %s %s: %s — %s", req.Method, req.URL.Path, resp.Status, strings.TrimSpace(string(raw)))
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			// try unmarshal into map for flexible responses
			var m map[string]interface{}
			if err2 := json.Unmarshal(raw, &m); err2 == nil {
				if outMap, ok := out.(*map[string]interface{}); ok {
					*outMap = m
					return nil
				}
			}
			return err
		}
	}
	return nil
}

func escapeSelector(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

func ServerURL(base string, deviceIndex, portBase, portStride int) string {
	if portBase <= 0 {
		portBase = 4723
	}
	if portStride <= 0 {
		portStride = 1
	}
	port := portBase + deviceIndex*portStride
	if base != "" && !strings.Contains(base, "4723") {
		// custom base without port substitution
		return strings.TrimRight(base, "/")
	}
	return fmt.Sprintf("http://127.0.0.1:%d", port)
}
