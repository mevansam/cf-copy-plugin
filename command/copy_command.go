package command

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/plugin"
	"github.com/mevansam/cf-copy-plugin/copy"
	"github.com/mevansam/cf-copy-plugin/helpers"
)

// CopyCommand -
type CopyCommand struct {
	sessionProvider helpers.CloudCountrollerSessionProvider
	targets         helpers.Targets

	o      *CopyOptions
	cli    plugin.CliConnection
	logger *helpers.Logger

	am copy.ApplicationsManager
	sm copy.ServicesManager

	srcOrg    models.OrganizationFields
	srcSpace  models.SpaceFields
	destOrg   models.OrganizationFields
	destSpace models.SpaceFields

	srcCCSession  helpers.CloudControllerSession
	destCCSession helpers.CloudControllerSession
}

// CopyOptions -
type CopyOptions struct {
	DestSpace         string
	DestOrg           string
	DestTarget        string
	SourceAppNames    []string
	AppHostFormat     string
	AppRouteDomain    string
	CopyAsDroplet     bool
	CopyAsUpsServices []string
	RecreateServices  bool
	ServicesOnly      bool

	Debug     bool
	TracePath string
}

// AppInfo
type appInfo struct {
	srcApp       models.Application
	dropletPath  string
	bindServices []string
}

// NewCopyCommand -
func NewCopyCommand(targets helpers.Targets, sessionProvider helpers.CloudCountrollerSessionProvider) CopyCmd {
	return &CopyCommand{sessionProvider: sessionProvider, targets: targets}
}

// Execute -
func (c *CopyCommand) Execute(cli plugin.CliConnection, o *CopyOptions) {

	defer c.cleanup()

	var (
		ok  bool
		err error
	)

	c.logger = helpers.NewLogger(o.Debug, o.TracePath)

	c.cli = cli
	c.o = o

	if ok, err = c.initialize(); ok {

		var (
			err     error
			message string

			am copy.ApplicationsManager
			ac copy.ApplicationCollection
			sm copy.ServicesManager
			sc copy.ServiceCollection
		)

		currentTarget, _ := c.targets.GetCurrentTarget()
		if currentTarget != o.DestTarget {
			message = fmt.Sprintf("Copying artifacts from %s %s / %s %s / %s %s to %s %s / %s %s / %s %s",
				terminal.HeaderColor("target"), terminal.EntityNameColor(currentTarget),
				terminal.HeaderColor("org"), terminal.EntityNameColor(c.srcOrg.Name),
				terminal.HeaderColor("space"), terminal.EntityNameColor(c.srcCCSession.GetSessionSpace().Name),
				terminal.HeaderColor("target"), terminal.EntityNameColor(c.o.DestTarget),
				terminal.HeaderColor("org"), terminal.EntityNameColor(c.o.DestOrg),
				terminal.HeaderColor("space"), terminal.EntityNameColor(c.o.DestSpace))
		} else if c.srcOrg.Name == c.o.DestOrg {
			message = fmt.Sprintf("Copying artifacts from %s %s / %s %s to %s %s / %s %s",
				terminal.HeaderColor("org"), terminal.EntityNameColor(c.srcOrg.Name),
				terminal.HeaderColor("space"), terminal.EntityNameColor(c.srcSpace.Name),
				terminal.HeaderColor("org"), terminal.EntityNameColor(c.o.DestOrg),
				terminal.HeaderColor("space"), terminal.EntityNameColor(c.o.DestSpace))
		} else {
			message = fmt.Sprintf("Copying artifacts %s %s to %s %s",
				terminal.HeaderColor("space"), terminal.EntityNameColor(c.srcSpace.Name),
				terminal.HeaderColor("space"), terminal.EntityNameColor(c.o.DestSpace))
		}
		message += fmt.Sprintf(" as %s...", terminal.EntityNameColor(c.destCCSession.GetSessionUsername()))
		c.logger.UI.Say(message)

		if currentTarget == o.DestTarget {
			// Restore source target on method exit. This needs
			// to be done when the source and destination targets
			// are the same. Otherwise the CLI session target
			// will be set to the destination target on exit.
			defer c.srcCCSession.SetSessionOrg(c.srcOrg)
			defer c.srcCCSession.SetSessionSpace(c.srcSpace)
		}

		am, err = copy.NewCfCliApplicationsManager(c.srcCCSession, c.destCCSession, c.logger)
		if err != nil {
			c.logger.UI.Failed(err.Error())
			return
		}
		defer am.Close()

		sm, err = copy.NewCfCliServicesManager(c.srcCCSession, c.destCCSession, o.DestTarget, o.DestOrg, o.DestSpace, c.logger)
		if err != nil {
			c.logger.UI.Failed(err.Error())
			return
		}
		defer sm.Close()

		if !o.ServicesOnly {
			ac, err = am.ApplicationsToBeCopied(o.SourceAppNames, o.CopyAsDroplet)
			if err != nil {
				c.logger.UI.Failed(err.Error())
				return
			}
		}

		sc, err = sm.ServicesToBeCopied(o.SourceAppNames, o.CopyAsUpsServices)
		if err != nil {
			c.logger.UI.Failed(err.Error())
			return
		}

		c.destCCSession.SetSessionOrg(c.destOrg)
		c.destCCSession.SetSessionSpace(c.destSpace)

		err = sm.DoCopy(sc, o.RecreateServices)
		if err != nil {
			c.logger.UI.Failed(err.Error())
			return
		}

		if !o.ServicesOnly {
			err = am.DoCopy(ac, sc, o.AppHostFormat, o.AppRouteDomain)
			if err != nil {
				c.logger.UI.Failed(err.Error())
				return
			}
		}

		c.logger.UI.Say("")
		c.logger.UI.Ok()
	}
	if err != nil {
		c.logger.UI.Failed(err.Error())
	}
}

func (c *CopyCommand) cleanup() {
	if c.srcCCSession != nil {
		c.srcCCSession.Close()
	}
	if c.destCCSession != nil {
		c.destCCSession.Close()
	}
}

func (c *CopyCommand) initialize() (ok bool, err error) {

	var (
		currentTarget string
		output        []string

		apps  []models.Application
		org   models.Organization
		space models.Space
	)

	if output, err = c.cli.CliCommandWithoutTerminalOutput("plugins"); err != nil {
		return
	}
	for _, s := range output[4:] {
		if len(s) > 0 && s[:10] == "cf-targets" {
			ok = true
			break
		}
	}
	if !ok {
		c.logger.UI.Failed("'Targets' plugin is requried to determine destination Cloud Foundry target.")
		return
	}

	ok = false
	if err = c.targets.Initialize(); err != nil {
		return
	}

	if currentTarget, err = c.targets.GetCurrentTarget(); err == nil {

		if c.o.DestTarget != "" {
			if c.o.DestTarget != currentTarget && !c.targets.HasTarget(c.o.DestTarget) {
				c.logger.UI.Failed("A target named '%s' cannot be found.", c.o.DestTarget)
				return
			}
		} else {
			c.o.DestTarget = currentTarget
		}

		// Initialize and validate source and destination sessions
		c.srcCCSession = c.sessionProvider.NewCloudControllerSessionFromFilepath(
			c.cli, c.targets.GetTargetConfigPath(currentTarget), c.logger)
		c.destCCSession = c.sessionProvider.NewCloudControllerSessionFromFilepath(
			c.cli, c.targets.GetTargetConfigPath(c.o.DestTarget), c.logger)

		if !c.srcCCSession.HasTarget() {
			c.logger.UI.Failed("The CLI target org and space needs to be set.")
			return
		}

		c.logger.DebugMessage("Options => %# v", c.o)
		c.logger.DebugMessage("Source Org => %# v\n", c.srcCCSession.GetSessionOrg())
		c.logger.DebugMessage("Source Space => %# v\n", c.srcCCSession.GetSessionSpace())

		if c.o.DestOrg == "" {
			c.o.DestOrg = c.srcCCSession.GetSessionOrg().Name
		}
		if currentTarget == c.o.DestTarget &&
			c.srcCCSession.GetSessionOrg().Name == c.o.DestOrg &&
			c.srcCCSession.GetSessionSpace().Name == c.o.DestSpace {

			c.logger.UI.Failed("The source and destination are the same.")
			return
		}

		apps, err = c.srcCCSession.AppSummary().GetSummariesInCurrentSpace()
		if err != nil {
			return
		}
		if len(c.o.SourceAppNames) == 0 {
			// Retrieve all application names to be copied
			for _, a := range apps {
				c.o.SourceAppNames = append(c.o.SourceAppNames, a.ApplicationFields.Name)
			}
		} else {
			// Validate source application exists
			for _, n := range c.o.SourceAppNames {
				if _, contains := helpers.ContainsApp(n, apps); !contains {
					if err != nil {
						c.logger.UI.Failed("The application '%s' does not exist.", n)
						return
					}
				}
			}
		}

		// Retrieve source and destination org and space
		c.srcOrg = c.srcCCSession.GetSessionOrg()
		c.srcSpace = c.srcCCSession.GetSessionSpace()

		org, err = c.destCCSession.Organizations().FindByName(c.o.DestOrg)
		if err != nil {
			return
		}
		c.destOrg = org.OrganizationFields

		space, err = c.destCCSession.Spaces().FindByNameInOrg(c.o.DestSpace, c.destOrg.GUID)
		if err != nil {
			return
		}
		c.destSpace = space.SpaceFields

		c.logger.DebugMessage("Destination Org => %# v", c.destOrg)
		c.logger.DebugMessage("Desintation Space => %# v", c.destSpace)

		ok = true
	}
	return
}
