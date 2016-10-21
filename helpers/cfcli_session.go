package helpers

import (
	"fmt"
	"os"
	"time"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/applications"
	"code.cloudfoundry.org/cli/cf/api/authentication"
	"code.cloudfoundry.org/cli/cf/api/organizations"
	"code.cloudfoundry.org/cli/cf/api/spaces"
	"code.cloudfoundry.org/cli/cf/api/strategy"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
)

// CfCliSessionProvider -
type CfCliSessionProvider struct{}

// CfCliSession -
type CfCliSession struct {
	logger *Logger

	config     coreconfig.Repository
	ccGateway  net.Gateway
	uaaGateway net.Gateway

	uaa authentication.UAARepository
}

// NewCfCliSessionProvider -
func NewCfCliSessionProvider() CloudCountrollerSessionProvider {
	return &CfCliSessionProvider{}
}

// NewCloudControllerSessionFromFilepath -
func (p *CfCliSessionProvider) NewCloudControllerSessionFromFilepath(configPath string, logger *Logger) CloudControllerSession {

	session := &CfCliSession{logger: logger}

	session.config = coreconfig.NewRepositoryFromFilepath(configPath, func(err error) {
		if err != nil {
			logger.UI.Failed(err.Error())
			os.Exit(1)
		}
	})

	if i18n.T == nil {
		i18n.T = i18n.Init(session.config.(i18n.LocalReader))
	}

	envDialTimeout := os.Getenv("CF_DIAL_TIMEOUT")
	session.ccGateway = net.NewCloudControllerGateway(session.config, time.Now, logger.UI, logger.TracePrinter, envDialTimeout)
	session.uaaGateway = net.NewUAAGateway(session.config, logger.UI, logger.TracePrinter, envDialTimeout)
	session.uaa = authentication.NewUAARepository(session.uaaGateway, session.config, net.NewRequestDumper(logger.TracePrinter))

	session.ccGateway.SetTokenRefresher(session.uaa)
	session.uaaGateway.SetTokenRefresher(session.uaa)

	return session
}

// Close -
func (s *CfCliSession) Close() {
	s.config.Close()
}

// HasTarget -
func (s *CfCliSession) HasTarget() bool {
	return s.config.HasOrganization() && s.config.HasSpace()
}

// GetSessionUsername -
func (s *CfCliSession) GetSessionUsername() string {
	return s.config.Username()
}

// GetSessionOrg -
func (s *CfCliSession) GetSessionOrg() models.OrganizationFields {
	return s.config.OrganizationFields()
}

// SetSessionOrg -
func (s *CfCliSession) SetSessionOrg(org models.OrganizationFields) {
	s.config.SetOrganizationFields(org)
}

// GetSessionSpace -
func (s *CfCliSession) GetSessionSpace() models.SpaceFields {
	return s.config.SpaceFields()
}

// SetSessionSpace -
func (s *CfCliSession) SetSessionSpace(space models.SpaceFields) {
	s.config.SetSpaceFields(space)
}

// Organizations -
func (s *CfCliSession) Organizations() organizations.OrganizationRepository {
	return organizations.NewCloudControllerOrganizationRepository(s.config, s.ccGateway)
}

// Spaces -
func (s *CfCliSession) Spaces() spaces.SpaceRepository {
	return spaces.NewCloudControllerSpaceRepository(s.config, s.ccGateway)
}

// Services -
func (s *CfCliSession) Services() api.ServiceRepository {
	return api.NewCloudControllerServiceRepository(s.config, s.ccGateway)
}

// ServicePlans -
func (s *CfCliSession) ServicePlans() api.ServicePlanRepository {
	return api.NewCloudControllerServicePlanRepository(s.config, s.ccGateway)
}

// ServiceSummary -
func (s *CfCliSession) ServiceSummary() api.ServiceSummaryRepository {
	return api.NewCloudControllerServiceSummaryRepository(s.config, s.ccGateway)
}

// UserProvidedServices -
func (s *CfCliSession) UserProvidedServices() api.UserProvidedServiceInstanceRepository {
	return api.NewCCUserProvidedServiceInstanceRepository(s.config, s.ccGateway)
}

// ServiceKeys -
func (s *CfCliSession) ServiceKeys() api.ServiceKeyRepository {
	return api.NewCloudControllerServiceKeyRepository(s.config, s.ccGateway)
}

// ServiceBindings -
func (s *CfCliSession) ServiceBindings() api.ServiceBindingRepository {
	return api.NewCloudControllerServiceBindingRepository(s.config, s.ccGateway)
}

// AppSummary -
func (s *CfCliSession) AppSummary() api.AppSummaryRepository {
	return api.NewCloudControllerAppSummaryRepository(s.config, s.ccGateway)
}

// Applications -
func (s *CfCliSession) Applications() applications.Repository {
	return applications.NewCloudControllerRepository(s.config, s.ccGateway)
}

// Routes -
func (s *CfCliSession) Routes() api.RouteRepository {
	return api.NewCloudControllerRouteRepository(s.config, s.ccGateway)
}

// Domains -
func (s *CfCliSession) Domains() api.DomainRepository {
	return api.NewCloudControllerDomainRepository(s.config, s.ccGateway, strategy.NewEndpointStrategy(s.config.APIVersion()))
}

// GetServiceCredentials -
func (s *CfCliSession) GetServiceCredentials(serviceBinding models.ServiceBindingFields) (*ServiceBindingDetail, error) {
	serviceBindingDetail := &ServiceBindingDetail{}
	url := fmt.Sprintf("%s"+serviceBinding.URL, s.config.APIEndpoint())
	err := s.ccGateway.GetResource(url, serviceBindingDetail)
	if err != nil {
		return nil, err
	}
	return serviceBindingDetail, nil
}

// UploadDroplet -
func (s *CfCliSession) UploadDroplet(appGUID string, contentType string, dropletUploadRequest *os.File) error {

	url := fmt.Sprintf("%s/v2/apps/%s/droplet/upload", s.config.APIEndpoint(), appGUID)
	request, err := s.ccGateway.NewRequestForFile("PUT", url, s.config.AccessToken(), dropletUploadRequest)
	if err != nil {
		return err
	}
	request.HTTPReq.Header.Set("Content-Type", contentType)

	response := make(map[string]interface{})
	_, err = s.ccGateway.PerformRequestForJSONResponse(request, &response)
	s.logger.DebugMessage("Response from droplet upload: #% v", response)

	return err
}
