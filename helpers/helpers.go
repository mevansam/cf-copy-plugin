package helpers

import (
	"time"

	"code.cloudfoundry.org/cli/cf/models"
)

// ContainsInStrings -
func ContainsInStrings(contains []string, strings []string) ([]string, bool) {
	stringsFound := []string{}
	for _, s := range strings {
		for _, ss := range contains {
			if s == ss {
				stringsFound = append(stringsFound, ss)
			}
		}
	}
	return stringsFound, len(stringsFound) > 0
}

// ContainsApp -
func ContainsApp(name string, apps []models.Application) (*models.Application, bool) {
	for _, a := range apps {
		if a.ApplicationFields.Name == name {
			return &a, true
		}
	}
	return nil, false
}

// ContainsService -
func ContainsService(name string, services []models.ServiceInstance) (*models.ServiceInstance, bool) {
	for _, s := range services {
		if s.Name == name {
			return &s, true
		}
	}
	return nil, false
}

// ContainsUserProvidedService -
func ContainsUserProvidedService(name string, userProvidedServices []models.UserProvidedService) (*models.UserProvidedService, bool) {
	for _, s := range userProvidedServices {
		if s.Name == name {
			return &s, true
		}
	}
	return nil, false
}

// ContainsServiceKey -
func ContainsServiceKey(name string, serviceKeys []models.ServiceKeyFields) (*models.ServiceKeyFields, bool) {
	for _, s := range serviceKeys {
		if s.Name == name {
			return &s, true
		}
	}
	return nil, false
}

// Retry -
func Retry(timeout time.Duration, poll time.Duration, callback func() (bool, error)) (err error) {

	var done bool

	timeoutAt := time.Now().Add(timeout * time.Millisecond)
	wait := poll * time.Millisecond

	for time.Now().Before(timeoutAt) {
		if done, err = callback(); done || err != nil {
			return
		}
		time.Sleep(wait)
	}
	return
}
