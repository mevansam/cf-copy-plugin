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
	downloadPath  string
	logger        *helpers.Logger
}

// CfCliApplicationCollection -
type CfCliApplicationCollection struct {
	defaultDomainAtSrc models.DomainFields
	applicationsToCopy []*Application
	appInfoMap         map[string]*Application
}

// NewCfCliApplicationsManager -
func NewCfCliApplicationsManager(
	srcCCSession helpers.CloudControllerSession,
	destCCSession helpers.CloudControllerSession,
	logger *helpers.Logger) (ApplicationsManager, error) {

	cliConfigPath, err := confighelpers.DefaultFilePath()
	if err != nil {
		return nil, err
	}
	downloadPath := filepath.Join(filepath.Dir(cliConfigPath), "appcontent")
	os.MkdirAll(downloadPath, os.ModePerm)

	return &CfCliApplicationsManager{
		srcCCSession:  srcCCSession,
		destCCSession: destCCSession,
		downloadPath:  downloadPath,
		logger:        logger,
	}, nil
}

// ApplicationsToBeCopied - Retrieve applications to copied
func (am *CfCliApplicationsManager) ApplicationsToBeCopied(
	appNames []string, copyAsDroplet bool) (ApplicationCollection, error) {

	am.logger.UI.Say("\nDownloading applications to be copied...")

	ac := &CfCliApplicationCollection{
		appInfoMap: make(map[string]*Application),
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

			app := NewApplication(a, am.downloadPath, copyAsDroplet)
			err = app.Content.Download(am.srcCCSession)
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

		am.logger.UI.Say("+ %s", terminal.EntityNameColor(a.srcApp.Name))

		destApp, err = am.destCCSession.Applications().Read(a.srcApp.Name)
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

		params := a.srcApp.ToParams()
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

		destApp, err = a.Content.Upload(am.destCCSession, params)
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
		err = helpers.Retry(60000, 5000, func() (cont bool, err error) {
			am.logger.DebugMessage("Waiting for applications bits of %s to finalize.", destApp.Name)
			app, err := am.destCCSession.AppSummary().GetSummary(destApp.GUID)
			if err != nil {
				return false, err
			}
			return app.PackageState == "STAGED", nil
		})

		am.logger.DebugMessage("Starting application %s.", destApp.Name)

		state = "started"
		_, err = am.destCCSession.Applications().Update(destApp.GUID, models.AppParams{State: &state})
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

		for _, r := range a.srcApp.Routes {

			am.logger.DebugMessage(
				"Using host part of route %s to source app to derive a unique route to the copied app.",
				r.URL())

			if appHostTmpl != nil {

				appHostVars["app"] = a.srcApp.Name
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
					err = am.destCCSession.Domains().ListDomainsForOrg(orgGUID, func(d models.DomainFields) bool {
						destDomainParts := strings.Split(d.Name, ".")
						if srcDomainParts[0] == destDomainParts[0] {
							domain = d
							found = true
						}
						return found
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
