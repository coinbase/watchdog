package types

// Component stands for datadog component type.
type Component string

var (
	// ComponentDashboard stands for dashboard or timeboard.
	ComponentDashboard = Component("dashboard")

	// ComponentMonitor stands for monitor.
	ComponentMonitor = Component("monitor")

	// ComponentScreenboard stands for screenboard.
	ComponentScreenboard = Component("screenboard")

	// ComponentDowntime stands for downtime.
	ComponentDowntime = Component("downtime")
)
