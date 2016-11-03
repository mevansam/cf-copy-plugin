package copy

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/cf/configuration/confighelpers"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/terminal"
	"github.com/mevansam/cf-copy-plugin/helpers"
)

// CfCliApplicationsManager -
type CfCliApplicationsManager struct {
	srcCCSession  helpers.CloudControllerSession
	destCCSession helpers.CloudControllerSession
	logger        *helpers.Logger

	copier       ApplicationCopier
	downloadPath string
}

// CfCliApplicationCollection -
type CfCliApplicationCollection struct {
	defaultDomainAtSrc models.DomainFields
	applicationsToCopy []ApplicationContent
	appInfoMap         map[string]ApplicationContent
}

// NewCfCliApplicationsManager -
func NewCfCliApplicationsManager() ApplicationsManager {
	return &CfCliApplicationsManager{copier: ApplicationCopier{}}
}

// Init -
func (am *CfCliApplicationsManager) Init(
	srcCCSession helpers.CloudControllerSession,
	destCCSession helpers.CloudControllerSession,
	logger *helpers.Logger) error {

	cliConfigPath, err := confighelpers.DefaultFilePath()
	if err != nil {
		return err
	}
	downloadPath := filepath.Join(filepath.Dir(cliConfigPath), "appcontent")
	os.RemoveAll(downloadPath)
	os.MkdirAll(downloadPath, os.ModePerm)

	am.srcCCSession = srcCCSession
	am.destCCSession = destCCSession
	am.downloadPath = downloadPath
	am.logger = logger

	return nil
}

// ApplicationsToBeCopied - Retrieve applications to copied
func (am *CfCliApplicationsManager) ApplicationsToBeCopied(
	appNames []string, copyAsDroplet bool) (ApplicationCollection, error) {

	am.logger.UI.Say("\nDownloading applications to be copied...")

	ac := &CfCliApplicationCollection{
		appInfoMap: make(map[string]ApplicationContent),
	}

	apps, err := am.srcCCSession.AppSummary().GetSummariesInCurrentSpace()
	if err != nil {
		return nil, err
	}

	for _, n := range appNames {
		if a, contains := helpers.ContainsApp(n, apps); contains {
			if err != nil {
				return nil, err
			}

			am.logger.UI.Say("+ downloading application %s", terminal.EntityNameColor(a.Name))

			app := am.copier.NewApplication(a, am.downloadPath, copyAsDroplet)
			err = app.Download(am.srcCCSession)
			if err != nil {
				return nil, err
			}
			ac.applicationsToCopy = append(ac.applicationsToCopy, app)
			ac.appInfoMap[a.Name] = app
		}
	}

	ac.defaultDomainAtSrc, err = am.srcCCSession.Domains().FirstOrDefault(am.srcCCSession.GetSessionOrg().GUID, nil)
	if err != nil {
		return nil, err
	}

	am.logger.DebugMessage("Applications to be copied => %# v", ac.applicationsToCopy)
	return ac, nil
}

// DoCopy -
func (am *CfCliApplicationsManager) DoCopy(
	applications ApplicationCollection, services ServiceCollection, appHostFormat string, appRouteDomain string) (err error) {

	var (
		appHostTmpl   *template.Template
		appHostVars   map[string]string
		appHostResult bytes.Buffer

		host  string
		route models.Route

		destApp models.Application
	)

	am.logger.UI.Say("\nCreating application copies at destination...")

	if len(appHostFormat) != 0 {
		appHostTmpl = template.Must(template.New("host").Parse(appHostFormat))

		appHostVars = make(map[string]string)
		appHostVars["org"] = am.destCCSession.GetSessionOrg().Name
		appHostVars["space"] = am.destCCSession.GetSessionSpace().Name
	}

	destSpace := am.destCCSession.GetSessionSpace()

	ac := applications.(*CfCliApplicationCollection)
	for _, a := range ac.applicationsToCopy {

		am.logger.UI.Say("+ %s", terminal.EntityNameColor(a.App().Name))

		destApp, err = am.destCCSession.Applications().Read(a.App().Name)
		if err != nil {
			if _, ok := err.(*errors.ModelNotFoundError); !ok {
				return
			}
		} else {
			for _, r := range destApp.Routes {
				am.logger.DebugMessage("Deleting existing application's route '%s'.", r.URL())
				err = am.destCCSession.Routes().Delete(r.GUID)
				if err != nil {
					return
				}
			}
			am.logger.DebugMessage("Deleting existing application to be copied '%s' at destination.", destApp.Name)
			err = am.destCCSession.Applications().Delete(destApp.GUID)
			if err != nil {
				return
			}
		}

		state := strings.ToUpper(models.ApplicationStateStopped)

		params := a.App().ToParams()
		params.GUID = nil
		params.State = &state
		params.BuildpackURL = nil
		params.DockerImage = nil
		params.StackGUID = nil
		params.SpaceGUID = &destSpace.GUID

		if params.HealthCheckType != nil && len(*params.HealthCheckType) == 0 {
			*params.HealthCheckType = "port"
		}

		am.logger.DebugMessage(
			"Uploading application %s using params: %# v",
			params.Name, params)

		destApp, err = a.Upload(am.destCCSession, params)
		if err != nil {
			return
		}

		if bindings, ok := services.AppBindings(destApp.Name); ok {
			for _, g := range bindings {
				am.logger.DebugMessage("Binding application %s to service %s.", destApp.Name, g)
				err = am.destCCSession.ServiceBindings().Create(g, destApp.GUID, make(map[string]interface{}))
				if err != nil {
					return
				}
			}
		}

		state = "started"
		params = models.AppParams{State: &state}

		err = helpers.Retry(300000, 5000, func() (bool, error) {
			am.logger.DebugMessage("Starting application %s.", destApp.Name)
			_, err = am.destCCSession.Applications().Update(destApp.GUID, models.AppParams{State: &state})
			if err != nil {
				am.logger.DebugMessage("Request to start application %s returned error: %s.", destApp.Name, err.Error())
				return false, err
			}
			return true, nil
		})
		if err != nil {
			return
		}

		var (
			orgGUID               string
			defaultDomain, domain models.DomainFields
		)

		orgGUID = am.destCCSession.GetSessionOrg().GUID
		defaultDomain, err = am.destCCSession.Domains().FirstOrDefault(orgGUID, nil)
		if err != nil {
			return
		}

		for _, r := range a.App().Routes {

			am.logger.DebugMessage(
				"Using host part of route %s to source app to derive a unique route to the copied app.",
				r.URL())

			if appHostTmpl != nil {

				appHostVars["app"] = a.App().Name
				appHostVars["host"] = r.Host

				appHostResult.Truncate(0)
				err = appHostTmpl.Execute(&appHostResult, appHostVars)
				if err != nil {
					return
				}
				host = appHostResult.String()
			} else {
				host = r.Host
			}

			if len(appRouteDomain) == 0 {

				// If route is on the default domain at source then create it on the
				// default domain at destination. Otherwise attempt to determine
				// the domain based on the first level of the source route's domain.

				domain = defaultDomain
				if r.Domain.Name != ac.defaultDomainAtSrc.Name {

					found := false
					srcDomainParts := strings.Split(r.Domain.Name, ".")
					am.logger.DebugMessage("Searching for match for source domain: %+v", srcDomainParts)

					err = am.destCCSession.Domains().ListDomainsForOrg(orgGUID, func(d models.DomainFields) bool {

						destDomainParts := strings.Split(d.Name, ".")
						am.logger.DebugMessage("Attempting to match dest domain: %+v", destDomainParts)

						if srcDomainParts[0] == destDomainParts[0] {
							domain = d
							found = true
						}
						return !found
					})
					if err != nil {
						return
					}
					if !found {
						am.logger.UI.Say(
							"  unable to create dest route to match source route %s",
							terminal.WarningColor(r.URL()))
						continue
					}
					am.logger.DebugMessage("Found matching dest domain: %# v", domain)
				}

			} else {
				domain, err = am.destCCSession.Domains().FirstOrDefault(orgGUID, &appRouteDomain)
			}

			route, err = am.destCCSession.Routes().Find(host, domain, "", 0)
			if err != nil {
				if _, ok := err.(*errors.ModelNotFoundError); !ok {
					return
				}
			} else {
				err = am.destCCSession.Routes().Delete(route.GUID)
				if err != nil {
					return
				}
			}
			route, err = am.destCCSession.Routes().Create(host, domain, "", 0, false)
			if err != nil {
				return
			}
			am.destCCSession.Routes().Bind(route.GUID, destApp.GUID)
			am.logger.UI.Say("  bound route %s", terminal.HeaderColor(route.URL()))
		}
	}
	return
}

// Close -
func (am *CfCliApplicationsManager) Close() {
	os.RemoveAll(am.downloadPath)
}
