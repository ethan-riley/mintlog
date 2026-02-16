package alerting

const (
	StateOK       = "ok"
	StateFiring   = "firing"
	StateResolved = "resolved"
)

// NextState determines the next alert state based on current state and whether threshold is exceeded.
func NextState(current string, exceeded bool) string {
	switch {
	case exceeded && current != StateFiring:
		return StateFiring
	case !exceeded && current == StateFiring:
		return StateOK
	default:
		return current
	}
}
