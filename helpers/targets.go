package helpers

// Targets -
type Targets interface {
	Initialize() error
	GetCurrentTarget() (string, error)
	HasTarget(target string) bool
	GetTargetConfigPath(target string) string
}
