package copy

import (
	"code.cloudfoundry.org/cli/cf/models"
	"github.com/mevansam/cf-copy-plugin/helpers"
)

// ApplicationsManager -
type ApplicationsManager interface {
	ApplicationsToBeCopied(appNames []string, copyAsDroplet bool) (ApplicationCollection, error)
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
	NewApplicationDroplet(downloadPath string) ApplicationContent
	NewApplicationBits(downloadPath string) ApplicationContent
}

// ApplicationContent -
type ApplicationContent interface {
	Download(session helpers.CloudControllerSession) error
	Upload(session helpers.CloudControllerSession, params models.AppParams) (models.Application, error)
}
