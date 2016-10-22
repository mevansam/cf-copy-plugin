package copy

import (
	"code.cloudfoundry.org/cli/cf/models"
	"github.com/mevansam/cf-copy-plugin/helpers"
)

// ApplicationsManager -
type ApplicationsManager interface {
	ApplicationsToBeCopied(appNames []string) (ApplicationCollection, error)
	DoCopy(applications ApplicationCollection, services ServiceCollection, appHostFormat string, appRouteDomain string) error
	Close()
}

// ApplicationCollection -
type ApplicationCollection interface {
}

// ServicesManager -
type ServicesManager interface {
	ServicesToBeCopied(appNames []string, upsServices []string) (ServiceCollection, error)
	DoCopy(services ServiceCollection, recreate bool) error
	Close()
}

// ServiceCollection -
type ServiceCollection interface {
	AppBindings(appName string) ([]string, bool)
}

// ApplicationContentProvider -
type ApplicationContentProvider interface {
	NewApplicationDroplet(session helpers.CloudControllerSession) ApplicationContent
	NewApplicationBits(session helpers.CloudControllerSession) ApplicationContent
}

// ApplicationContent -
type ApplicationContent interface {
	Download(app models.Application) error
	Upload(params models.AppParams) error
}
