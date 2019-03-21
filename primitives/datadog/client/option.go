package client

// Option is a functional parameter for datadog client.
type Option func(c *Client) error

// WithRemoveDashboardFields sets fields to be removed from a dashboard response.
// Nested fields are supported via comma for example (dash.modified) will remove the modified field from
// a nested dict under "dash"
// endpoint /api/v1/dash/<id>
func WithRemoveDashboardFields(fields []string) Option {
	return func(c *Client) error {
		c.removeDashboardFields = fields
		return nil
	}
}

// WithRemoveMonitorFields sets fields to be removed from monitor response.
// endpoint /api/v1/monitor/<id>
func WithRemoveMonitorFields(monitorFields, alertFields []string) Option {
	return func(c *Client) error {
		c.removeMonitorFields = monitorFields
		c.removeAlertFields = alertFields
		return nil
	}
}

// WithRemoveScreenBoardFields sets fields to be removed from screen board response.
// endpoint /api/v1/screen/<id>
func WithRemoveScreenBoardFields(fields []string) Option {
	return func(c *Client) error {
		c.removeScreenboardFields = fields
		return nil
	}
}
