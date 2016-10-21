package copy

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
	DoCopy(services ServiceCollection) error
	Close()
}

// ServiceCollection -
type ServiceCollection interface {
	AppBindings(appName string) ([]string, bool)
}
