package copy

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
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

// CfCliApplicationsManager -
type CfCliApplicationsManager struct {
	srcCCSession  helpers.CloudControllerSession
	destCCSession helpers.CloudControllerSession

	dropletDownloader   droplet.Droplet
	dropletDownloadPath string

	logger *helpers.Logger
}

// CfCliApplicationCollection -
type CfCliApplicationCollection struct {
	applicationsToCopy []*appInfo
	appInfoMap         map[string]*appInfo
}

// appInfo
type appInfo struct {
	srcApp       models.Application
	dropletPath  string
	bindServices []string
}

// NewCfCliApplicationsManager -
func NewCfCliApplicationsManager(
	srcCCSession helpers.CloudControllerSession,
	destCCSession helpers.CloudControllerSession,
	cli plugin.CliConnection,
	logger *helpers.Logger) (ApplicationsManager, error) {

	cliConfigPath, err := confighelpers.DefaultFilePath()
	if err != nil {
		return nil, err
	}
	dropletDownloadPath := filepath.Join(filepath.Dir(cliConfigPath), "droplets")
	dropletDownloader := droplet.NewCFDroplet(cli, &droplet.CFDownloader{
		Cli:    cli,
		Writer: new(droplet.CFFileWriter),
	})

	return &CfCliApplicationsManager{
		srcCCSession:        srcCCSession,
		destCCSession:       destCCSession,
		dropletDownloader:   dropletDownloader,
		dropletDownloadPath: dropletDownloadPath,
		logger:              logger,
	}, nil
}

// ApplicationsToBeCopied - Retrieve applications to copied
func (am *CfCliApplicationsManager) ApplicationsToBeCopied(
	appNames []string) (ApplicationCollection, error) {

	am.logger.UI.Say("\nDownloading droplets of applications that will be copied...")

	ac := &CfCliApplicationCollection{
		appInfoMap: make(map[string]*appInfo),
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
			info := appInfo{
				srcApp:       *a,
				dropletPath:  am.downloadDroplet(n),
				bindServices: []string{},
			}
			ac.applicationsToCopy = append(ac.applicationsToCopy, &info)
			ac.appInfoMap[a.Name] = &info
		}
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
			"Uploading droplet of applcation %s at %s using params: %# v",
			destApp.Name, a.dropletPath, params)

		destApp, err = am.uploadDroplet(params, a.dropletPath)
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
		err = helpers.RetryOnError(2, 3, func() (err error) {
			am.logger.DebugMessage("Starting application %s.", destApp.Name)
			_, err = am.destCCSession.Applications().Update(destApp.GUID, models.AppParams{State: &state})
			if err != nil {
				am.logger.DebugMessage("Unable to start application: %s", err.Error())
			}
			return
		})

		var domain models.DomainFields
		if len(appRouteDomain) == 0 {
			domain, err = am.destCCSession.Domains().FirstOrDefault(am.destCCSession.GetSessionOrg().GUID, nil)
		} else {
			domain, err = am.destCCSession.Domains().FirstOrDefault(am.destCCSession.GetSessionOrg().GUID, &appRouteDomain)
		}

		for _, r := range a.srcApp.Routes {

			am.logger.DebugMessage(
				"Using host part of route %s to source app to derive a unique route to the copied app.",
				r.URL())

			if appHostTmpl != nil {

				appHostVars["app"] = a.srcApp.Name
				appHostVars["host"] = r.Host

				err = appHostTmpl.Execute(&appHostResult, appHostVars)
				if err != nil {
					return
				}
				host = appHostResult.String()
			} else {
				host = r.Host
			}
			route, err = am.destCCSession.Routes().Create(host, domain, "", -1, false)
			if err != nil {
				return
			}
			am.destCCSession.Routes().Bind(route.GUID, destApp.GUID)
			am.logger.UI.Say("- bound route %s", terminal.HeaderColor(route.URL()))
		}
	}
	return
}

// Close -
func (am *CfCliApplicationsManager) Close() {
	// os.RemoveAll(am.dropletDownloadPath)
}

func (am *CfCliApplicationsManager) downloadDroplet(app string) string {

	// Download application droplet
	am.logger.UI.Say("+ downloading droplet for application %s", terminal.EntityNameColor(app))

	dropletAppPath := filepath.Join(filepath.Join(am.dropletDownloadPath, app), app) + ".tgz"
	am.logger.DebugMessage("Downloading droplet '%s'.", dropletAppPath)

	// os.MkdirAll(filepath.Dir(dropletAppPath), os.ModePerm)
	// am.dropletDownloader.SaveDroplet(app, dropletAppPath)
	return dropletAppPath
}

func (am *CfCliApplicationsManager) uploadDroplet(params models.AppParams, dropletPath string) (app models.Application, err error) {

	app, err = am.destCCSession.Applications().Create(params)
	if err != nil {
		return
	}

	dropletUploadRequest, err := ioutil.TempFile("", ".droplet")
	if err != nil {
		return
	}
	file, err := os.Open(dropletPath)
	if err != nil {
		return
	}
	defer func() {
		file.Close()
		dropletUploadRequest.Close()
		os.Remove(dropletUploadRequest.Name())
	}()

	writer := multipart.NewWriter(dropletUploadRequest)
	part, err := writer.CreateFormFile("droplet", filepath.Base(dropletPath))
	if err != nil {
		return
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return
	}
	err = writer.Close()
	if err != nil {
		return
	}

	am.destCCSession.UploadDroplet(app.GUID, writer.FormDataContentType(), dropletUploadRequest)
	return
}
