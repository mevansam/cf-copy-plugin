package helpers

import (
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/applications"
	"code.cloudfoundry.org/cli/cf/api/organizations"
	"code.cloudfoundry.org/cli/cf/api/spaces"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
)

// CloudCountrollerSessionProvider -
type CloudCountrollerSessionProvider interface {
	NewCloudControllerSessionFromFilepath(configPath string, ui terminal.UI, logger trace.Printer) CloudControllerSession
}

// CloudControllerSession -
type CloudControllerSession interface {
	Close()

	HasTarget() bool

	GetSessionUsername() string
	GetSessionOrg() models.OrganizationFields
	SetSessionOrg(models.OrganizationFields)
	GetSessionSpace() models.SpaceFields
	SetSessionSpace(models.SpaceFields)

	// Cloud Countroller APIs

	Organizations() organizations.OrganizationRepository
	Spaces() spaces.SpaceRepository

	ServiceSummary() api.ServiceSummaryRepository
	Services() api.ServiceRepository
	UserProvidedServices() api.UserProvidedServiceInstanceRepository
	ServiceKeys() api.ServiceKeyRepository

	AppSummary() api.AppSummaryRepository
	Applications() applications.Repository

	ServiceBindings() api.ServiceBindingRepository
	GetServiceCredentials(models.ServiceBindingFields) (*ServiceBindingDetail, error)
}

// Model structs not present in CF CLI API

// ServiceBindingDetail -
type ServiceBindingDetail struct {
	Entity struct {
		AppGUID             string                 `json:"entapp_guidity,omitempty"`
		ServiceInstanceGUID string                 `json:"service_instance_guid,omitempty"`
		Credentials         map[string]interface{} `json:"credentials,omitempty"`
	} `json:"entity,omitempty"`
}
