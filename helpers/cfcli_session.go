package helpers

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/mitchellh/ioprogress"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/applicationbits"
	"code.cloudfoundry.org/cli/cf/api/applications"
	"code.cloudfoundry.org/cli/cf/api/authentication"
	"code.cloudfoundry.org/cli/cf/api/organizations"
	"code.cloudfoundry.org/cli/cf/api/spaces"
	"code.cloudfoundry.org/cli/cf/api/strategy"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/plugin"
)

// CfCliSessionProvider -
type CfCliSessionProvider struct{}

// CfCliSession -
type CfCliSession struct {
	logger *Logger

	config     coreconfig.Repository
	ccGateway  net.Gateway
	uaaGateway net.Gateway

	httpClient *http.Client

	uaa authentication.UAARepository
}

// NewCfCliSessionProvider -
func NewCfCliSessionProvider() CloudCountrollerSessionProvider {
	return &CfCliSessionProvider{}
}

// NewCloudControllerSessionFromFilepath -
func (p *CfCliSessionProvider) NewCloudControllerSessionFromFilepath(
	cli plugin.CliConnection,
	configPath string,
	logger *Logger) CloudControllerSession {

	sslDisabled, _ := cli.IsSSLDisabled()

	session := &CfCliSession{
		logger: logger,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: sslDisabled},
			},
		},
	}

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

// ApplicationBits -
func (s *CfCliSession) ApplicationBits() applicationbits.Repository {
	return applicationbits.NewCloudControllerApplicationBitsRepository(s.config, s.ccGateway)
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

// DownloadAppContent -
func (s *CfCliSession) DownloadAppContent(appGUID string, outputFile *os.File, asDroplet bool) (err error) {

	var url string
	if asDroplet {
		url = fmt.Sprintf("%s/v2/apps/%s/droplet/download", s.config.APIEndpoint(), appGUID)
	} else {
		url = fmt.Sprintf("%s/v2/apps/%s/download", s.config.APIEndpoint(), appGUID)
	}
	request, err := s.ccGateway.NewRequest("GET", url, s.config.AccessToken(), nil)
	if err != nil {
		return
	}

	response, err := s.httpClient.Do(request.HTTPReq)
	if err != nil {
		if _, ok := err.(*errors.InvalidTokenError); !ok {
			// Handle token refresh error
			var newToken string
			newToken, err = s.uaa.RefreshAuthToken()
			if err == nil {
				request.HTTPReq.Header.Set("Authorization", newToken)
				response, err = s.httpClient.Do(request.HTTPReq)
			}
		}
		if err != nil {
			return
		}
	}
	defer response.Body.Close()
	progressReader := &ioprogress.Reader{
		Reader:   response.Body,
		Size:     response.ContentLength,
		DrawFunc: ioprogress.DrawTerminalf(os.Stdout, drawProgressBar()),
	}
	_, err = io.Copy(outputFile, progressReader)
	return
}

// UploadDroplet -
func (s *CfCliSession) UploadDroplet(appGUID string, contentType string, dropletUploadRequest *os.File) error {

	fileStats, err := dropletUploadRequest.Stat()
	if err != nil {
		return err
	}
	fileSize := fileStats.Size()

	progressReader := readSeekerWrapper{
		seeker: dropletUploadRequest,
		reader: &ioprogress.Reader{
			Reader:   dropletUploadRequest,
			Size:     fileSize,
			DrawFunc: ioprogress.DrawTerminalf(os.Stdout, drawProgressBar()),
		},
	}
	_, _ = progressReader.Seek(0, 0)

	url := fmt.Sprintf("%s/v2/apps/%s/droplet/upload", s.config.APIEndpoint(), appGUID)
	request, err := s.ccGateway.NewRequest("PUT", url, s.config.AccessToken(), progressReader)
	if err != nil {
		return err
	}
	request.HTTPReq.Header.Set("Content-Type", contentType)
	request.HTTPReq.ContentLength = fileSize

	response := make(map[string]interface{})
	_, err = s.ccGateway.PerformRequestForJSONResponse(request, &response)
	s.logger.DebugMessage("Response from droplet upload: %# v", response)

	return err
}

// drawProgressBar -
func drawProgressBar() ioprogress.DrawTextFormatFunc {

	bar := ioprogress.DrawTextFormatBar(60)
	return func(progress, total int64) string {
		return fmt.Sprintf(
			"  %s %s",
			bar(progress, total),
			ioprogress.DrawTextFormatBytes(progress, total))
	}
}

// readSeakerWrapper -
type readSeekerWrapper struct {
	seeker io.ReadSeeker
	reader io.Reader
}

func (r readSeekerWrapper) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r readSeekerWrapper) Seek(offset int64, whence int) (next int64, err error) {
	return r.seeker.Seek(offset, whence)
}
