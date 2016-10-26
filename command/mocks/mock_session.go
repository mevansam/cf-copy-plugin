package mock_test

import (
	"os"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/applicationbits"
	"code.cloudfoundry.org/cli/cf/api/applications"
	"code.cloudfoundry.org/cli/cf/api/organizations"
	"code.cloudfoundry.org/cli/cf/api/spaces"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/plugin"
	"github.com/mevansam/cf-copy-plugin/helpers"
)

// MockSession -
type MockSession struct {
	MockHasTarget            func() bool
	MockGetSessionUsername   func() string
	MockGetSessionOrg        func() models.OrganizationFields
	MockSetSessionOrg        func(models.OrganizationFields)
	MockGetSessionSpace      func() models.SpaceFields
	MockSetSessionSpace      func(models.SpaceFields)
	MockOrganizations        func() organizations.OrganizationRepository
	MockSpaces               func() spaces.SpaceRepository
	MockServices             func() api.ServiceRepository
	MockServicePlans         func() api.ServicePlanRepository
	MockServiceSummary       func() api.ServiceSummaryRepository
	MockUserProvidedServices func() api.UserProvidedServiceInstanceRepository
	MockServiceKeys          func() api.ServiceKeyRepository
	MockServiceBindings      func() api.ServiceBindingRepository
	MockAppSummary           func() api.AppSummaryRepository
	MockApplications         func() applications.Repository
	MockApplicationBits      func() applicationbits.Repository
	MockRoutes               func() api.RouteRepository
	MockDomains              func() api.DomainRepository

	MockGetServiceCredentials func(models.ServiceBindingFields) (*helpers.ServiceBindingDetail, error)
	MockDownloadAppContent    func(string, *os.File, bool) error
	MockUploadDroplet         func(string, string, *os.File) error
}

// NewCloudControllerSessionFromFilepath -
func (m *MockSession) NewCloudControllerSessionFromFilepath(
	cli plugin.CliConnection,
	configPath string,
	logger *helpers.Logger) helpers.CloudControllerSession {

	return m
}

// Close -
func (m *MockSession) Close() {
}

// HasTarget -
func (m *MockSession) HasTarget() bool {
	return m.MockHasTarget()
}

// GetSessionUsername -
func (m *MockSession) GetSessionUsername() string {
	return m.MockGetSessionUsername()
}

// GetSessionOrg -
func (m *MockSession) GetSessionOrg() models.OrganizationFields {
	return m.MockGetSessionOrg()
}

// SetSessionOrg -
func (m *MockSession) SetSessionOrg(org models.OrganizationFields) {
	m.MockSetSessionOrg(org)
}

// GetSessionSpace -
func (m *MockSession) GetSessionSpace() models.SpaceFields {
	return m.MockGetSessionSpace()
}

// SetSessionSpace -
func (m *MockSession) SetSessionSpace(space models.SpaceFields) {
	m.MockSetSessionSpace(space)
}

// Organizations -
func (m *MockSession) Organizations() organizations.OrganizationRepository {
	return m.MockOrganizations()
}

// Spaces -
func (m *MockSession) Spaces() spaces.SpaceRepository {
	return m.MockSpaces()
}

// Services -
func (m *MockSession) Services() api.ServiceRepository {
	return m.MockServices()
}

// ServicePlans -
func (m *MockSession) ServicePlans() api.ServicePlanRepository {
	return m.MockServicePlans()
}

// ServiceSummary -
func (m *MockSession) ServiceSummary() api.ServiceSummaryRepository {
	return m.MockServiceSummary()
}

// UserProvidedServices -
func (m *MockSession) UserProvidedServices() api.UserProvidedServiceInstanceRepository {
	return m.MockUserProvidedServices()
}

// ServiceKeys -
func (m *MockSession) ServiceKeys() api.ServiceKeyRepository {
	return m.MockServiceKeys()
}

// AppSummary -
func (m *MockSession) AppSummary() api.AppSummaryRepository {
	return m.AppSummary()
}

// Applications -
func (m *MockSession) Applications() applications.Repository {
	return m.MockApplications()
}

// ApplicationBits -
func (m *MockSession) ApplicationBits() applicationbits.Repository {
	return m.MockApplicationBits()
}

// Routes -
func (m *MockSession) Routes() api.RouteRepository {
	return m.MockRoutes()
}

// Domains -
func (m *MockSession) Domains() api.DomainRepository {
	return m.MockDomains()
}

// ServiceBindings -
func (m *MockSession) ServiceBindings() api.ServiceBindingRepository {
	return m.MockServiceBindings()
}

// GetServiceCredentials -
func (m *MockSession) GetServiceCredentials(serviceBinding models.ServiceBindingFields) (*helpers.ServiceBindingDetail, error) {
	return m.MockGetServiceCredentials(serviceBinding)
}

// DownloadAppContent -
func (m *MockSession) DownloadAppContent(appGUID string, outputFile *os.File, asDroplet bool) error {
	return m.MockDownloadAppContent(appGUID, outputFile, asDroplet)
}

// UploadDroplet -
func (m *MockSession) UploadDroplet(appGUID string, contentType string, dropletUploadRequest *os.File) error {
	return m.MockUploadDroplet(appGUID, contentType, dropletUploadRequest)
}
