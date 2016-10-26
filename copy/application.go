package copy

import (
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/mevansam/cf-copy-plugin/helpers"

	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/models"
)

// Application -
type Application struct {
	srcApp       models.Application
	bindServices []string

	Content ApplicationContent
}

// AppBits -
type AppBits struct {
	appGUID     string
	filePath    string
	extractPath string
}

// AppDroplet -
type AppDroplet struct {
	appGUID  string
	filePath string
}

// NewApplication -
func NewApplication(srcApp *models.Application, downloadPath string, copyAsDroplet bool) *Application {

	app := &Application{
		srcApp: *srcApp,
	}

	if copyAsDroplet {
		app.Content = app.NewApplicationDroplet(downloadPath)
	} else {
		app.Content = app.NewApplicationBits(downloadPath)
	}

	return app
}

// NewApplicationDroplet -
func (a *Application) NewApplicationDroplet(downloadPath string) ApplicationContent {
	return &AppDroplet{
		appGUID:  a.srcApp.GUID,
		filePath: filepath.Join(downloadPath, a.srcApp.Name) + ".tgz",
	}
}

// NewApplicationBits -
func (a *Application) NewApplicationBits(downloadPath string) ApplicationContent {
	return &AppBits{
		appGUID:     a.srcApp.GUID,
		filePath:    filepath.Join(downloadPath, a.srcApp.Name) + ".zip",
		extractPath: filepath.Join(downloadPath, a.srcApp.Name),
	}
}

// Download -
func (b *AppBits) Download(session helpers.CloudControllerSession) (err error) {
	outputFile, err := os.OpenFile(b.filePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	defer outputFile.Close()
	if err != nil {
		return
	}

	err = session.DownloadAppContent(b.appGUID, outputFile, false)
	if err != nil {
		return
	}
	return
}

// Upload -
func (b *AppBits) Upload(session helpers.CloudControllerSession, params models.AppParams) (app models.Application, err error) {

	app, err = session.Applications().Create(params)
	if err != nil {
		return
	}

	file, err := os.Open(b.filePath)
	if err != nil {
		return
	}
	defer file.Close()

	err = session.ApplicationBits().UploadBits(app.GUID, file, []resources.AppFileResource{})
	if err != nil {
		return
	}
	return
}

// Download -
func (d *AppDroplet) Download(session helpers.CloudControllerSession) (err error) {
	outputFile, err := os.OpenFile(d.filePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	defer outputFile.Close()
	if err != nil {
		return
	}
	err = session.DownloadAppContent(d.appGUID, outputFile, true)
	return
}

// Upload -
func (d *AppDroplet) Upload(session helpers.CloudControllerSession, params models.AppParams) (app models.Application, err error) {

	app, err = session.Applications().Create(params)
	if err != nil {
		return
	}

	dropletUploadRequest, err := ioutil.TempFile("", ".droplet")
	if err != nil {
		return
	}
	file, err := os.Open(d.filePath)
	if err != nil {
		return
	}
	defer func() {
		file.Close()
		dropletUploadRequest.Close()
		os.Remove(dropletUploadRequest.Name())
	}()

	writer := multipart.NewWriter(dropletUploadRequest)
	part, err := writer.CreateFormFile("droplet", filepath.Base(d.filePath))
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

	session.UploadDroplet(app.GUID, writer.FormDataContentType(), dropletUploadRequest)
	return
}
