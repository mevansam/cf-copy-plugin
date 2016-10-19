package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/cf/configuration/confighelpers"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/plugin"
	"github.com/krujos/download_droplet_plugin/droplet"
	"github.com/mevansam/cf-copy-plugin/helpers"
)

// CopyCommand -
type CopyCommand struct {
	logger *helpers.Logger
	cli    plugin.CliConnection

	o *CopyOptions

	sessionProvider helpers.CloudCountrollerSessionProvider
	srcCCSession    helpers.CloudControllerSession
	destCCSession   helpers.CloudControllerSession

	srcOrg    models.OrganizationFields
	srcSpace  models.SpaceFields
	destOrg   models.OrganizationFields
	destSpace models.SpaceFields

	dropletDownloader   droplet.Droplet
	dropletDownloadPath string

	targets helpers.Targets
}

// CopyOptions -
type CopyOptions struct {
	DestSpace         string
	DestOrg           string
	DestTarget        string
	SourceAppNames    []string
	AppHostFormat     string
	AppRouteDomain    string
	CopyAsUpsServices []string
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
			message string

			applicationsToCopy []*appInfo
			appInfoMap         map[string]*appInfo

			serviceInstance        models.ServiceInstance
			serviceInstancesToCopy []models.ServiceInstance

			destUserProvidedServices []models.UserProvidedService
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
		// defer os.RemoveAll(c.dropletDownloadPath)

		// Retrieve applications to copied

		applicationsToCopy = []*appInfo{}
		if !c.o.ServicesOnly {
			c.logger.UI.Say("\nDownloading droplets of applications that will be copied...")

			appInfoMap = make(map[string]*appInfo)
			for _, n := range o.SourceAppNames {
				a, err := c.srcCCSession.Applications().Read(n)
				if err != nil {
					c.logger.UI.Failed(err.Error())
					return
				}
				info := appInfo{
					srcApp:       a,
					dropletPath:  c.downloadDroplet(n),
					bindServices: []string{},
				}
				applicationsToCopy = append(applicationsToCopy, &info)
				appInfoMap[a.ApplicationFields.Name] = &info
			}
		}

		// Retrieve details of service instances to be copied

		upsSummaries, err := c.srcCCSession.UserProvidedServices().GetSummaries()
		if err != nil {
			c.logger.UI.Failed(err.Error())
			return
		}
		upsServices := []models.UserProvidedService{}
		for _, u := range upsSummaries.Resources {
			upsServices = append(upsServices, u.UserProvidedService)
		}

		services, err := c.srcCCSession.ServiceSummary().GetSummariesInCurrentSpace()
		if err != nil {
			c.logger.UI.Failed(err.Error())
			return
		}
		for _, s := range services {
			serviceInstance, err = c.srcCCSession.Services().FindInstanceByName(s.ServiceInstanceFields.Name)
			if err != nil {
				c.logger.UI.Failed(err.Error())
				return
			}

			if appNames, contains := helpers.ContainsInStrings(o.SourceAppNames, s.ServiceInstanceFields.ApplicationNames); contains {

				serviceInstance.ApplicationNames = appNames
				serviceInstancesToCopy = append(serviceInstancesToCopy, serviceInstance)

				for _, name := range appNames {
					if info, ok := appInfoMap[name]; ok {
						info.bindServices = append(info.bindServices, serviceInstance.ServiceInstanceFields.Name)
					}
				}

				keyName := fmt.Sprintf("__%s_copy_for_/%s/%s/%s",
					serviceInstance.ServiceInstanceFields.Name, o.DestTarget, o.DestOrg, o.DestSpace)
				if serviceKey, contains := helpers.ContainsServiceKey(keyName, serviceInstance.ServiceKeys); contains {
					c.logger.DebugMessage("Deleting existing service key %s for service %s.", keyName, serviceInstance.ServiceInstanceFields.Name)
					c.srcCCSession.ServiceKeys().DeleteServiceKey(serviceKey.GUID)
				}

				if ups, contains := helpers.ContainsUserProvidedService(serviceInstance.ServiceInstanceFields.Name, upsServices); contains &&
					len(serviceInstance.ServicePlan.GUID) == 0 && len(serviceInstance.ServiceOffering.GUID) == 0 {

					c.logger.DebugMessage("User provided service '%s' to copy: %# v",
						serviceInstance.ServiceInstanceFields.Name, serviceInstance)
					destUserProvidedServices = append(destUserProvidedServices, *ups)

				} else {

					// Managed services copied as a user-provided-service in the target space
					// will use credentials from a service key created in the source space.

					if _, contains := helpers.ContainsInStrings([]string{serviceInstance.ServiceInstanceFields.Name}, o.CopyAsUpsServices); contains {

						c.logger.DebugMessage("Managed service '%s' that will be copied as a user provided service: %# v",
							serviceInstance.ServiceInstanceFields.Name, serviceInstance)

						c.logger.DebugMessage(
							"Creating service key %s for service %s to be used as source of credentials for target user-provided-service.",
							keyName, serviceInstance.ServiceInstanceFields.Name)

						c.srcCCSession.ServiceKeys().CreateServiceKey(serviceInstance.ServiceInstanceFields.GUID, keyName, make(map[string]interface{}))

						serviceKey, err := c.srcCCSession.ServiceKeys().GetServiceKey(serviceInstance.ServiceInstanceFields.GUID, keyName)
						if err != nil {
							c.logger.UI.Failed(err.Error())
							return
						}

						c.logger.DebugMessage("Service key created for copying managed service as a user provided service: %# v", serviceKey)

						ups := models.UserProvidedService{
							Name:        serviceInstance.ServiceInstanceFields.Name,
							Credentials: serviceKey.Credentials,
						}
						destUserProvidedServices = append(destUserProvidedServices, ups)
					} else {
						c.logger.DebugMessage("Managed service '%s' that will be re-created as a managed service at the destination: %# v",
							serviceInstance.ServiceInstanceFields.Name, serviceInstance)
					}
				}
			}
		}

		c.logger.DebugMessage("Applications to be copied => %# v", applicationsToCopy)
		c.logger.DebugMessage("Services to be copied => %# v", serviceInstancesToCopy)

		c.destCCSession.SetSessionOrg(c.destOrg)
		c.destCCSession.SetSessionSpace(c.destSpace)

		// Create service instance copies at destination

		c.logger.UI.Say("\nCreating service copies at destination...")

		servicesAtDest, err := c.destCCSession.ServiceSummary().GetSummariesInCurrentSpace()
		if err != nil {
			c.logger.UI.Failed(err.Error())
			return
		}
		for _, s := range serviceInstancesToCopy {

			rebindAppGUIDS := []string{}

			if _, contains := helpers.ContainsService(s.ServiceInstanceFields.Name, servicesAtDest); contains {

				serviceInstance, err = c.destCCSession.Services().FindInstanceByName(s.ServiceInstanceFields.Name)
				if err != nil {
					c.logger.UI.Failed(err.Error())
					return
				}

				c.logger.DebugMessage("Unbinding any applications bound to service instance %s at destination.", s.ServiceInstanceFields.Name)
				for _, appName := range serviceInstance.ApplicationNames {
					app, err := c.destCCSession.Applications().Read(appName)
					if err != nil {
						if _, ok := err.(*errors.ModelNotFoundError); !ok {
							c.logger.UI.Failed(err.Error())
							return
						}
					} else {
						c.destCCSession.ServiceBindings().Delete(serviceInstance, app.ApplicationFields.GUID)
						rebindAppGUIDS = append(rebindAppGUIDS, app.ApplicationFields.GUID)
					}
				}

				c.logger.DebugMessage("Deleting all service keys of service instance %s at destination.", s.ServiceInstanceFields.Name)
				for _, serviceKey := range serviceInstance.ServiceKeys {
					c.destCCSession.ServiceKeys().DeleteServiceKey(serviceKey.GUID)
				}

				c.logger.DebugMessage("Deleting existing service instance %s at destination.", s.ServiceInstanceFields.Name)
				err = c.destCCSession.Services().DeleteService(serviceInstance)
				if err != nil {
					c.logger.UI.Failed(err.Error())
					return
				}
			}
			if ups, contains := helpers.ContainsUserProvidedService(s.ServiceInstanceFields.Name, destUserProvidedServices); contains {

				c.logger.UI.Say("  + %s as a user provided service instance at destination",
					terminal.EntityNameColor(s.ServiceInstanceFields.Name))

				err = c.destCCSession.UserProvidedServices().Create(ups.Name, "", "", ups.Credentials)
				if err != nil {
					c.logger.UI.Failed(err.Error())
					return
				}
				c.logger.DebugMessage("Created user provided service %s at destination.", s.ServiceInstanceFields.Name)

			} else {
				c.logger.UI.Say("  + %s as a managed service instance at destination",
					terminal.EntityNameColor(s.ServiceInstanceFields.Name))

				err = c.destCCSession.Services().CreateServiceInstance(s.ServiceInstanceFields.Name,
					s.ServicePlan.GUID, s.ServiceInstanceFields.Params, s.ServiceInstanceFields.Tags)
				if err != nil {
					c.logger.UI.Failed(err.Error())
					return
				}
				c.logger.DebugMessage("Created managed service %s at destination.", s.ServiceInstanceFields.Name)
			}

			serviceInstance, err = c.destCCSession.Services().FindInstanceByName(s.ServiceInstanceFields.Name)
			if err != nil {
				c.logger.UI.Failed(err.Error())
				return
			}
			for _, g := range rebindAppGUIDS {
				c.logger.DebugMessage("Rebinding app with GUID %s to service %s.", g, serviceInstance.ServiceInstanceFields.Name)
				err = c.destCCSession.ServiceBindings().Create(serviceInstance.ServiceInstanceFields.GUID, g, make(map[string]interface{}))
				if err != nil {
					c.logger.UI.Failed(err.Error())
					return
				}
			}
		}

		if !c.o.ServicesOnly {
			c.logger.UI.Say("\nCreating application copies at destination...")

			for _, a := range applicationsToCopy {
				c.logger.UI.Say("  + %s", terminal.EntityNameColor(a.srcApp.ApplicationFields.Name))

				destApp, err := c.destCCSession.Applications().Read(a.srcApp.ApplicationFields.Name)
				if err != nil {
					if _, ok := err.(*errors.ModelNotFoundError); !ok {
						c.logger.UI.Failed(err.Error())
						return
					}
				} else {
					c.logger.DebugMessage("Deleting existing application to be copied '%s' at destination.", destApp.ApplicationFields.Name)
					err = c.destCCSession.Applications().Delete(destApp.ApplicationFields.GUID)
					if err != nil {
						c.logger.UI.Failed(err.Error())
					}
				}

				state := strings.ToUpper(models.ApplicationStateStopped)

				params := a.srcApp.ToParams()
				params.State = &state
				params.BuildpackURL = nil
				params.DockerImage = nil
				params.SpaceGUID = &c.destSpace.GUID
				params.ServicesToBind = a.bindServices

				destApp, err = c.destCCSession.Applications().Create(params)
				if err != nil {
					c.logger.UI.Failed(err.Error())
					return
				}
				c.logger.DebugMessage("Created new application copy: %# v", destApp)
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

		cfPath string
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
			c.targets.GetTargetConfigPath(currentTarget), c.logger.UI, c.logger.TracePrinter)
		c.destCCSession = c.sessionProvider.NewCloudControllerSessionFromFilepath(
			c.targets.GetTargetConfigPath(c.o.DestTarget), c.logger.UI, c.logger.TracePrinter)

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

		cfPath, err = confighelpers.DefaultFilePath()
		if err != nil {
			c.logger.UI.Failed(err.Error())
			return
		}
		c.dropletDownloadPath = filepath.Join(filepath.Dir(cfPath), "droplets")
		c.dropletDownloader = droplet.NewCFDroplet(c.cli, &droplet.CFDownloader{
			Cli:    c.cli,
			Writer: new(droplet.CFFileWriter),
		})

		ok = true
	}
	return
}

func (c *CopyCommand) downloadDroplet(app string) string {

	// Download application droplet
	c.logger.UI.Say("  + download application %s", terminal.EntityNameColor(app))

	dropletAppPath := filepath.Join(filepath.Join(c.dropletDownloadPath, app), app) + ".tgz"
	c.logger.DebugMessage("Downloading droplet '%s'.", dropletAppPath)

	os.MkdirAll(filepath.Dir(dropletAppPath), os.ModePerm)
	c.dropletDownloader.SaveDroplet(app, dropletAppPath)
	return dropletAppPath
}
