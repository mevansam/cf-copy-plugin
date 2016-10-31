package mock_test

import (
	"github.com/mevansam/cf-copy-plugin/copy"
	"github.com/mevansam/cf-copy-plugin/helpers"
)

// MockServicesManager -
type MockServicesManager struct {
	MockInit func(srcCCSession helpers.CloudControllerSession,
		destCCSession helpers.CloudControllerSession,
		destTarget, destOrg, destSpace string,
		logger *helpers.Logger) error

	MockServicesToBeCopied func(appNames []string,
		upsServices []string) (copy.ServiceCollection, error)

	MockDoCopy func(services copy.ServiceCollection, recreate bool) error

	MockClose func()
}

// Close -
func (m *MockServicesManager) Close() {
	if m.MockClose != nil {
		m.MockClose()
	}
}

// Init -
func (m *MockServicesManager) Init(srcCCSession helpers.CloudControllerSession,
	destCCSession helpers.CloudControllerSession,
	destTarget, destOrg, destSpace string,
	logger *helpers.Logger) (err error) {

	if m.MockInit != nil {
		err = m.MockInit(srcCCSession, destCCSession, destTarget, destOrg, destSpace, logger)
	}
	return
}

// ServicesToBeCopied -
func (m *MockServicesManager) ServicesToBeCopied(appNames []string,
	upsServices []string) (sc copy.ServiceCollection, err error) {

	if m.MockServicesToBeCopied != nil {
		sc, err = m.MockServicesToBeCopied(appNames, upsServices)
	}
	return
}

// DoCopy -
func (m *MockServicesManager) DoCopy(services copy.ServiceCollection, recreate bool) (err error) {

	if m.MockDoCopy != nil {
		err = m.MockDoCopy(services, recreate)
	}
	return err
}
