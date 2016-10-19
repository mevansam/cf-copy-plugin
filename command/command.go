package command

import "code.cloudfoundry.org/cli/plugin"

// CopyCmd - Provides IoC for the Copy Implementation
type CopyCmd interface {
	Execute(cli plugin.CliConnection, o *CopyOptions)
}
