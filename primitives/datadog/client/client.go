package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	datadogAPI = "https://api.datadoghq.com/api/v1"
)

const (
	apiKeyField = "api_key"
	appKeyField = "application_key"
)

// Component stands for datadog component.
type Component string

const (
	monitorType     = Component("monitor")
	dashboardType   = Component("dash")
	screenboardType = Component("screen")
	alertType       = Component("alert")
	downtimeType    = Component("downtime")
)

var (
	// ErrInvalidDashboard is returned if the passed dashboard object is invalid.
	ErrInvalidDashboard = errors.New("invalid dashboard")

	// ErrInvalidMonitor is returned if the passed monitor object is invalid.
	ErrInvalidMonitor = errors.New("invalid monitor")

	// ErrInvalidDowntime is returned if the passed downtime object is invalid.
	ErrInvalidDowntime = errors.New("invalid downtime")

	// ErrInvalidScreenboard is returned if the passed screen board object is invalid.
	ErrInvalidScreenboard = errors.New("invalid screenboard")

	// ErrInvalidAlert s returned if the passed alert object in invalid.
	ErrInvalidAlert = errors.New("invalid alert")

	// errInvalidComponent is a generic error returned by internal function.
	errInvalidComponent = errors.New("invalid component")
)

// New returns an datadog client.
func New(apiKey, appKey string, opts ...Option) (*Client, error) {

	c := &Client{
		apiKey:       apiKey,
		appKey:       appKey,
		baseEndpoint: datadogAPI,

		httpClient: &http.Client{},
	}

	for _, opt := range opts {
		if opt != nil {
			err := opt(c)
			if err != nil {
				return nil, err
			}
		}
	}

	return c, nil
}

// Client represents a datadog API.
type Client struct {
	baseEndpoint string
	apiKey       string
	appKey       string
	httpClient   *http.Client

	removeDashboardFields   []string
	removeMonitorFields     []string
	removeAlertFields       []string
	removeScreenboardFields []string
}

func (c Client) do(method, apiCall string, b io.Reader) ([]byte, error) {
	values := url.Values{}
	values.Add(apiKeyField, c.apiKey)
	values.Add(appKeyField, c.appKey)

	apiURL := strings.Join([]string{c.baseEndpoint, apiCall}, "/")
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse URL: %s", apiURL)
	}

	u.RawQuery = values.Encode()

	req, err := http.NewRequest(method, u.String(), b)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create a new request, URL: %s", apiURL)
	}

	logrus.Debugf("[%s] %s", req.Method, req.URL.Path)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to make a request, URL: %s", apiURL)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read response, URL: %s", apiURL)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return nil, errors.Errorf("invalid status code %d, URL %s, method: %s, Response: %s", resp.StatusCode, apiURL, method, string(body))
	}

	return body, nil
}

func (c Client) genericUpdate(component Component, monitor json.RawMessage) error {
	m := struct {
		ID int `json:"id"`
	}{}

	err := json.Unmarshal(monitor, &m)
	if err != nil {
		return errors.Wrapf(err, "unable to unmarshal %s", component)
	}

	if m.ID == 0 {
		return errInvalidComponent
	}

	_, err = c.do("PUT", fmt.Sprintf("%s/%d", component, m.ID), bytes.NewReader(monitor))
	if err != nil {
		return err
	}

	return nil
}

func (c Client) stripJSONFields(body []byte, fields []string) ([]byte, error) {

	if len(fields) == 0 {
		return body, nil
	}

	container, err := gabs.ParseJSON(body)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse json before stripping field %s", fields)
	}

	for _, field := range fields {
		err = container.Delete(strings.Split(field, ".")...)
		if err != nil {
			logrus.Errorf("unable to strip fields [%v] from json %s: %s", fields, container, err)
		}
	}
	return container.EncodeJSON(gabs.EncodeOptHTMLEscape(false)), nil
}
