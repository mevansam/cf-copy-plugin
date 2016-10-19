package helpers

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/cf/configuration/confighelpers"
)

// TargetsPluginInfo -
type TargetsPluginInfo struct {
	targetsPath  string
	targetSuffix string

	targetConfigPaths map[string]string
}

// NewTargetsPluginInfo -
func NewTargetsPluginInfo() Targets {

	defaultFilePath, _ := confighelpers.DefaultFilePath()

	t := TargetsPluginInfo{
		targetsPath:       filepath.Join(filepath.Dir(defaultFilePath), "targets"),
		targetSuffix:      "." + filepath.Base(defaultFilePath),
		targetConfigPaths: make(map[string]string),
	}

	return &t
}

// Initialize -
func (t *TargetsPluginInfo) Initialize() error {

	var (
		err   error
		files []os.FileInfo
	)

	if files, err = ioutil.ReadDir(t.targetsPath); err != nil {
		return err
	}
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, t.targetSuffix) {
			t.targetConfigPaths[strings.TrimSuffix(filename, t.targetSuffix)] = filepath.Join(t.targetsPath, filename)
		}
	}
	return nil
}

// GetCurrentTarget -
func (t *TargetsPluginInfo) GetCurrentTarget() (string, error) {
	currentPath, err := filepath.EvalSymlinks(filepath.Join(t.targetsPath, "current"))
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(filepath.Base(currentPath), t.targetSuffix), nil
}

// HasTarget -
func (t *TargetsPluginInfo) HasTarget(target string) bool {
	_, exists := t.targetConfigPaths[target]
	return exists
}

// GetTargetConfigPath -
func (t *TargetsPluginInfo) GetTargetConfigPath(target string) (path string) {

	if currentTarget, _ := t.GetCurrentTarget(); currentTarget == target {
		path, _ = confighelpers.DefaultFilePath()
	} else {
		path = t.targetConfigPaths[target]
	}
	return
}
