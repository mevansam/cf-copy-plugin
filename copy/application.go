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

// ApplicationCopier -
type ApplicationCopier struct {
}

// Application -
type Application struct {
	srcApp       *models.Application
	bindServices []string
}

// AppBits -
type AppBits struct {
	Application

	appGUID     string
	filePath    string
	extractPath string
}

// AppDroplet -
type AppDroplet struct {
	Application

	appGUID  string
	filePath string
}

// NewApplication -
func (a *ApplicationCopier) NewApplication(srcApp *models.Application, downloadPath string, copyAsDroplet bool) (app ApplicationContent) {

	if copyAsDroplet {
		appDroplet := &AppDroplet{}
		appDroplet.srcApp = srcApp
		appDroplet.filePath = filepath.Join(downloadPath, srcApp.Name) + ".tgz"
		app = appDroplet
	} else {
		appBits := &AppBits{}
		appBits.srcApp = srcApp
		appBits.filePath = filepath.Join(downloadPath, srcApp.Name) + ".zip"
		appBits.extractPath = filepath.Join(downloadPath, srcApp.Name)
		app = appBits
	}
	return
}

// App -
func (b *AppBits) App() *models.Application {
	return b.srcApp
}

// Download -
func (b *AppBits) Download(session helpers.CloudControllerSession) (err error) {
	outputFile, err := os.OpenFile(b.filePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	defer outputFile.Close()
	if err != nil {
		return
	}

	err = session.DownloadAppContent(b.srcApp.GUID, outputFile, false)
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

// App -
func (d *AppDroplet) App() *models.Application {
	return d.srcApp
}

// Download -
func (d *AppDroplet) Download(session helpers.CloudControllerSession) (err error) {
	outputFile, err := os.OpenFile(d.filePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	defer outputFile.Close()
	if err != nil {
		return
	}
	err = session.DownloadAppContent(d.srcApp.GUID, outputFile, true)
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
