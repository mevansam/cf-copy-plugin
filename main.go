package main

import (
	"github.com/mevansam/cf-copy-plugin/command"
	"github.com/mevansam/cf-copy-plugin/copy"
	"github.com/mevansam/cf-copy-plugin/helpers"
)

func main() {

	// Copy Command using CF CLI API
	c := command.NewCopyCommand(
		helpers.NewTargetsPluginInfo(),
		helpers.NewCfCliSessionProvider(),
		copy.NewCfCliApplicationsManager(),
		copy.NewCfCliServicesManager())

	command.NewCopyPlugin(c).Start()
}
