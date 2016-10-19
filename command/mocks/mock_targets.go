package mock_test

// MockTargets -
type MockTargets struct {
	CurrentTarget string
	Targets       map[string]string
}

// Initialize -
func (t *MockTargets) Initialize() error {
	return nil
}

// GetCurrentTarget -
func (t *MockTargets) GetCurrentTarget() (string, error) {
	return t.CurrentTarget, nil
}

// HasTarget -
func (t *MockTargets) HasTarget(target string) bool {
	_, ok := t.Targets[target]
	return ok
}

// GetTargetConfigPath -
func (t *MockTargets) GetTargetConfigPath(target string) string {
	return t.Targets[target]
}
