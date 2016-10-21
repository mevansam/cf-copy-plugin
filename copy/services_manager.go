package copy

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/terminal"
	"github.com/mevansam/cf-copy-plugin/helpers"
)

// CfCliServicesManager -
type CfCliServicesManager struct {
	srcCCSession  helpers.CloudControllerSession
	destCCSession helpers.CloudControllerSession
	logger        *helpers.Logger

	serviceKeyFormat string
}

// CfCliServiceCollection -
type CfCliServiceCollection struct {
	serviceInstancesToCopy []models.ServiceInstance

	destServiceInstanceMap   map[string]models.ServiceInstance
	destAppBindings          map[string][]string
	destUserProvidedServices []models.UserProvidedService
}

// NewCfCliServicesManager -
func NewCfCliServicesManager(
	srcCCSession helpers.CloudControllerSession,
	destCCSession helpers.CloudControllerSession,
	destTarget, destOrg, destSpace string,
	logger *helpers.Logger) (ServicesManager, error) {

	return &CfCliServicesManager{

		srcCCSession:  srcCCSession,
		destCCSession: destCCSession,
		logger:        logger,

		serviceKeyFormat: "__%s_copy_for_" + fmt.Sprintf("/%s/%s/%s", destTarget, destOrg, destSpace),
	}, nil
}

// ServicesToBeCopied - Retrieve details of service instances to be copied
func (sm *CfCliServicesManager) ServicesToBeCopied(
	appNames []string, copyAsUpsServices []string) (ServiceCollection, error) {

	sc := &CfCliServiceCollection{
		destServiceInstanceMap: make(map[string]models.ServiceInstance),
		destAppBindings:        make(map[string][]string),
	}

	upsSummaries, err := sm.srcCCSession.UserProvidedServices().GetSummaries()
	if err != nil {
		return nil, err
	}
	upsServices := []models.UserProvidedService{}
	for _, u := range upsSummaries.Resources {
		upsServices = append(upsServices, u.UserProvidedService)
	}

	services, err := sm.srcCCSession.ServiceSummary().GetSummariesInCurrentSpace()
	if err != nil {
		return nil, err
	}
	for _, s := range services {
		serviceInstance, err := sm.srcCCSession.Services().FindInstanceByName(s.Name)
		if err != nil {
			return nil, err
		}

		if appNames, contains := helpers.ContainsInStrings(appNames, s.ApplicationNames); contains {

			serviceInstance.ApplicationNames = appNames
			sc.serviceInstancesToCopy = append(sc.serviceInstancesToCopy, serviceInstance)

			keyName := fmt.Sprintf(sm.serviceKeyFormat, serviceInstance.Name)
			if serviceKey, contains := helpers.ContainsServiceKey(keyName, serviceInstance.ServiceKeys); contains {
				sm.logger.DebugMessage("Deleting existing service key %s for service %s.", keyName, serviceInstance.Name)
				sm.srcCCSession.ServiceKeys().DeleteServiceKey(serviceKey.GUID)
			}

			if ups, contains := helpers.ContainsUserProvidedService(serviceInstance.Name, upsServices); contains &&
				len(serviceInstance.ServicePlan.GUID) == 0 && len(serviceInstance.ServiceOffering.GUID) == 0 {

				sm.logger.DebugMessage("User provided service '%s' to copy: %# v",
					serviceInstance.Name, serviceInstance)
				sc.destUserProvidedServices = append(sc.destUserProvidedServices, *ups)

			} else {

				// Managed services copied as a user-provided-service in the target space
				// will use credentials from a service key created in the source space.

				if _, contains := helpers.ContainsInStrings([]string{serviceInstance.Name}, copyAsUpsServices); contains {

					sm.logger.DebugMessage("Managed service '%s' that will be copied as a user provided service: %# v",
						serviceInstance.Name, serviceInstance)

					sm.logger.DebugMessage(
						"Creating service key %s for service %s to be used as source of credentials for target user-provided-service.",
						keyName, serviceInstance.Name)

					sm.srcCCSession.ServiceKeys().CreateServiceKey(serviceInstance.GUID, keyName, make(map[string]interface{}))

					serviceKey, err := sm.srcCCSession.ServiceKeys().GetServiceKey(serviceInstance.GUID, keyName)
					if err != nil {
						return nil, err
					}

					sm.logger.DebugMessage("Service key created for copying managed service as a user provided service: %# v", serviceKey)

					ups := models.UserProvidedService{
						Name:        serviceInstance.Name,
						Credentials: serviceKey.Credentials,
					}
					sc.destUserProvidedServices = append(sc.destUserProvidedServices, ups)
				} else {
					sm.logger.DebugMessage("Managed service '%s' that will be re-created as a managed service at the destination: %# v",
						serviceInstance.Name, serviceInstance)
				}
			}
		}
	}

	sm.logger.DebugMessage("Services to be copied => %# v", sc.serviceInstancesToCopy)
	return sc, nil
}

// DoCopy - Create service instance copies at destination
func (sm *CfCliServicesManager) DoCopy(services ServiceCollection) (err error) {

	var (
		ok bool
	)

	sm.logger.UI.Say("\nCreating service copies at destination...")

	sc := (services).(*CfCliServiceCollection)

	servicesAtDest, err := sm.destCCSession.ServiceSummary().GetSummariesInCurrentSpace()
	if err != nil {
		return
	}
	for _, s := range sc.serviceInstancesToCopy {

		var (
			serviceInstance models.ServiceInstance
			rebindAppGUIDS  []string
			offerings       models.ServiceOfferings
			plans           []models.ServicePlanFields
		)

		if _, contains := helpers.ContainsService(s.Name, servicesAtDest); contains {

			serviceInstance, err = sm.destCCSession.Services().FindInstanceByName(s.Name)
			if err != nil {
				return
			}

			sm.logger.DebugMessage(
				"Found service instance having the same name as service to be copied: %# v",
				serviceInstance)

			for _, binding := range serviceInstance.ServiceBindings {

				sm.logger.DebugMessage(
					"Unbinding application with GUID %s bound to service instance %s at destination.",
					binding.AppGUID, serviceInstance.Name)

				sm.destCCSession.ServiceBindings().Delete(serviceInstance, binding.AppGUID)
				rebindAppGUIDS = append(rebindAppGUIDS, binding.AppGUID)
			}

			for _, serviceKey := range serviceInstance.ServiceKeys {

				sm.logger.DebugMessage(
					"Deleting all service key with GUID %s of service instance %s at destination.",
					serviceKey.GUID, serviceInstance.Name)

				sm.destCCSession.ServiceKeys().DeleteServiceKey(serviceKey.GUID)
			}

			serviceInstance.ServiceBindings = []models.ServiceBindingFields{}
			serviceInstance.ServiceKeys = []models.ServiceKeyFields{}

			err = helpers.RetryOnError(2, 3, func() (err error) {
				sm.logger.DebugMessage("Deleting existing service instance %s at destination.", serviceInstance.Name)
				err = sm.destCCSession.Services().DeleteService(serviceInstance)
				if err != nil {
					sm.logger.DebugMessage("Unable to delete service instance: %s", err.Error())
				}
				return
			})
			if err != nil {
				return
			}
		}
		if ups, contains := helpers.ContainsUserProvidedService(s.Name, sc.destUserProvidedServices); contains {

			sm.logger.UI.Say("+ %s as a user provided service instance at destination",
				terminal.EntityNameColor(s.Name))

			err = sm.destCCSession.UserProvidedServices().Create(ups.Name, "", "", ups.Credentials)
			if err != nil {
				return
			}
			sm.logger.DebugMessage("Created user provided service %s at destination.", s.Name)

		} else {
			sm.logger.UI.Say("+ %s as a managed service instance at destination",
				terminal.EntityNameColor(s.Name))

			sm.logger.DebugMessage("Debug looking up the GUID for service '%s' plan name '%s'",
				s.ServiceOffering.Label, s.ServicePlan.Name)

			offerings, err = sm.destCCSession.Services().FindServiceOfferingsForSpaceByLabel(
				sm.destCCSession.GetSessionSpace().GUID, s.ServiceOffering.Label)
			if err != nil {
				return
			}

			servicePlanGUID := ""
			for _, o := range offerings {
				plans, err = sm.destCCSession.ServicePlans().Search(map[string]string{"service_guid": o.GUID})
				if err != nil {
					return
				}
				for _, p := range plans {
					if p.Name == s.ServicePlan.Name {
						servicePlanGUID = p.GUID
					}
				}
			}
			if servicePlanGUID == "" {
				err = fmt.Errorf("Unable to determine the GUID for service '%s' plan name '%s'",
					s.ServiceOffering.Label, s.Name)
				return
			}

			sm.logger.DebugMessage("GUID for service '%s' plan name '%s' is: %s",
				s.ServiceOffering.Label, s.Name, servicePlanGUID)

			err = sm.destCCSession.Services().CreateServiceInstance(s.Name,
				servicePlanGUID, s.Params, s.Tags)
			if err != nil {
				return
			}
			sm.logger.DebugMessage("Created managed service %s at destination.", s.Name)
		}

		serviceInstance, err = sm.destCCSession.Services().FindInstanceByName(s.Name)
		if err != nil {
			return
		}
		sc.destServiceInstanceMap[serviceInstance.Name] = serviceInstance

		for _, g := range rebindAppGUIDS {
			sm.logger.DebugMessage("Rebinding app with GUID %s to service %s.", g, serviceInstance.Name)
			err = sm.destCCSession.ServiceBindings().Create(serviceInstance.GUID, g, make(map[string]interface{}))
			if err != nil {
				return
			}
		}

		for _, a := range s.ApplicationNames {
			if _, ok = sc.destAppBindings[a]; ok {
				sc.destAppBindings[a] = append(sc.destAppBindings[a], serviceInstance.GUID)
			} else {
				sc.destAppBindings[a] = append([]string{}, serviceInstance.GUID)
			}
		}
	}
	return nil
}

// Close -
func (sm *CfCliServicesManager) Close() {
}

// AppBindings -
func (sc CfCliServiceCollection) AppBindings(appName string) (bindings []string, ok bool) {
	bindings, ok = sc.destAppBindings[appName]
	return
}
