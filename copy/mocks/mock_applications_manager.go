package mock_test

import (
	"github.com/mevansam/cf-copy-plugin/copy"
	"github.com/mevansam/cf-copy-plugin/helpers"
)

// MockApplicationsManager -
type MockApplicationsManager struct {
	MockInit func(srcCCSession helpers.CloudControllerSession,
		destCCSession helpers.CloudControllerSession,
		logger *helpers.Logger) error

	MockApplicationsToBeCopied func(appNames []string,
		copyAsDroplet bool) (copy.ApplicationCollection, error)

	MockDoCopy func(applications copy.ApplicationCollection,
		services copy.ServiceCollection,
		appHostFormat string,
		appRouteDomain string) error

	MockClose func()
}

// Close -
func (m *MockApplicationsManager) Close() {
	if m.MockClose != nil {
		m.MockClose()
	}
}

// Init -
func (m *MockApplicationsManager) Init(srcCCSession helpers.CloudControllerSession,
	destCCSession helpers.CloudControllerSession,
	logger *helpers.Logger) (err error) {

	if m.MockInit != nil {
		err = m.MockInit(srcCCSession, destCCSession, logger)
	}
	return
}

// ApplicationsToBeCopied -
func (m *MockApplicationsManager) ApplicationsToBeCopied(appNames []string,
	copyAsDroplet bool) (ac copy.ApplicationCollection, err error) {

	if m.MockApplicationsToBeCopied != nil {
		ac, err = m.MockApplicationsToBeCopied(appNames, copyAsDroplet)
	}
	return
}

// DoCopy -
func (m *MockApplicationsManager) DoCopy(applications copy.ApplicationCollection,
	services copy.ServiceCollection,
	appHostFormat string,
	appRouteDomain string) (err error) {

	if m.MockDoCopy != nil {
		err = m.MockDoCopy(applications, services, appHostFormat, appRouteDomain)
	}
	return nil
}
