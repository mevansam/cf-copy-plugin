package main

import (
	"github.com/mevansam/cf-cli-api/cfapi"
	"github.com/mevansam/cf-cli-api/copy"
	"github.com/mevansam/cf-copy-plugin/command"
	"github.com/mevansam/cf-copy-plugin/helpers"
)

func main() {

	// Copy Command using CF CLI API
	c := command.NewCopyCommand(
		helpers.NewTargetsPluginInfo(),
		cfapi.NewCfCliSessionProvider(),
		copy.NewCfCliApplicationsManager(),
		copy.NewCfCliServicesManager())

	command.NewCopyPlugin(c).Start()
}
