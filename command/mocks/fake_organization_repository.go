package mock_test

import (
	"sync"

	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/api/organizations"
)

type FakeOrganizationRepository struct {
	ListOrgsStub        func(limit int) ([]models.Organization, error)
	listOrgsMutex       sync.RWMutex
	listOrgsArgsForCall []struct {
		limit int
	}
	listOrgsReturns struct {
		result1 []models.Organization
		result2 error
	}
	GetManyOrgsByGUIDStub        func(orgGUIDs []string) (orgs []models.Organization, apiErr error)
	getManyOrgsByGUIDMutex       sync.RWMutex
	getManyOrgsByGUIDArgsForCall []struct {
		orgGUIDs []string
	}
	getManyOrgsByGUIDReturns struct {
		result1 []models.Organization
		result2 error
	}
	FindByNameStub        func(name string) (org models.Organization, apiErr error)
	findByNameMutex       sync.RWMutex
	findByNameArgsForCall []struct {
		name string
	}
	findByNameReturns struct {
		result1 models.Organization
		result2 error
	}
	CreateStub        func(org models.Organization) (apiErr error)
	createMutex       sync.RWMutex
	createArgsForCall []struct {
		org models.Organization
	}
	createReturns struct {
		result1 error
	}
	RenameStub        func(orgGUID string, name string) (apiErr error)
	renameMutex       sync.RWMutex
	renameArgsForCall []struct {
		orgGUID string
		name    string
	}
	renameReturns struct {
		result1 error
	}
	DeleteStub        func(orgGUID string) (apiErr error)
	deleteMutex       sync.RWMutex
	deleteArgsForCall []struct {
		orgGUID string
	}
	deleteReturns struct {
		result1 error
	}
	SharePrivateDomainStub        func(orgGUID string, domainGUID string) (apiErr error)
	sharePrivateDomainMutex       sync.RWMutex
	sharePrivateDomainArgsForCall []struct {
		orgGUID    string
		domainGUID string
	}
	sharePrivateDomainReturns struct {
		result1 error
	}
	UnsharePrivateDomainStub        func(orgGUID string, domainGUID string) (apiErr error)
	unsharePrivateDomainMutex       sync.RWMutex
	unsharePrivateDomainArgsForCall []struct {
		orgGUID    string
		domainGUID string
	}
	unsharePrivateDomainReturns struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeOrganizationRepository) ListOrgs(limit int) ([]models.Organization, error) {
	fake.listOrgsMutex.Lock()
	fake.listOrgsArgsForCall = append(fake.listOrgsArgsForCall, struct {
		limit int
	}{limit})
	fake.recordInvocation("ListOrgs", []interface{}{limit})
	fake.listOrgsMutex.Unlock()
	if fake.ListOrgsStub != nil {
		return fake.ListOrgsStub(limit)
	} else {
		return fake.listOrgsReturns.result1, fake.listOrgsReturns.result2
	}
}

func (fake *FakeOrganizationRepository) ListOrgsCallCount() int {
	fake.listOrgsMutex.RLock()
	defer fake.listOrgsMutex.RUnlock()
	return len(fake.listOrgsArgsForCall)
}

func (fake *FakeOrganizationRepository) ListOrgsArgsForCall(i int) int {
	fake.listOrgsMutex.RLock()
	defer fake.listOrgsMutex.RUnlock()
	return fake.listOrgsArgsForCall[i].limit
}

func (fake *FakeOrganizationRepository) ListOrgsReturns(result1 []models.Organization, result2 error) {
	fake.ListOrgsStub = nil
	fake.listOrgsReturns = struct {
		result1 []models.Organization
		result2 error
	}{result1, result2}
}

func (fake *FakeOrganizationRepository) GetManyOrgsByGUID(orgGUIDs []string) (orgs []models.Organization, apiErr error) {
	var orgGUIDsCopy []string
	if orgGUIDs != nil {
		orgGUIDsCopy = make([]string, len(orgGUIDs))
		copy(orgGUIDsCopy, orgGUIDs)
	}
	fake.getManyOrgsByGUIDMutex.Lock()
	fake.getManyOrgsByGUIDArgsForCall = append(fake.getManyOrgsByGUIDArgsForCall, struct {
		orgGUIDs []string
	}{orgGUIDsCopy})
	fake.recordInvocation("GetManyOrgsByGUID", []interface{}{orgGUIDsCopy})
	fake.getManyOrgsByGUIDMutex.Unlock()
	if fake.GetManyOrgsByGUIDStub != nil {
		return fake.GetManyOrgsByGUIDStub(orgGUIDs)
	} else {
		return fake.getManyOrgsByGUIDReturns.result1, fake.getManyOrgsByGUIDReturns.result2
	}
}

func (fake *FakeOrganizationRepository) GetManyOrgsByGUIDCallCount() int {
	fake.getManyOrgsByGUIDMutex.RLock()
	defer fake.getManyOrgsByGUIDMutex.RUnlock()
	return len(fake.getManyOrgsByGUIDArgsForCall)
}

func (fake *FakeOrganizationRepository) GetManyOrgsByGUIDArgsForCall(i int) []string {
	fake.getManyOrgsByGUIDMutex.RLock()
	defer fake.getManyOrgsByGUIDMutex.RUnlock()
	return fake.getManyOrgsByGUIDArgsForCall[i].orgGUIDs
}

func (fake *FakeOrganizationRepository) GetManyOrgsByGUIDReturns(result1 []models.Organization, result2 error) {
	fake.GetManyOrgsByGUIDStub = nil
	fake.getManyOrgsByGUIDReturns = struct {
		result1 []models.Organization
		result2 error
	}{result1, result2}
}

func (fake *FakeOrganizationRepository) FindByName(name string) (org models.Organization, apiErr error) {
	fake.findByNameMutex.Lock()
	fake.findByNameArgsForCall = append(fake.findByNameArgsForCall, struct {
		name string
	}{name})
	fake.recordInvocation("FindByName", []interface{}{name})
	fake.findByNameMutex.Unlock()
	if fake.FindByNameStub != nil {
		return fake.FindByNameStub(name)
	} else {
		return fake.findByNameReturns.result1, fake.findByNameReturns.result2
	}
}

func (fake *FakeOrganizationRepository) FindByNameCallCount() int {
	fake.findByNameMutex.RLock()
	defer fake.findByNameMutex.RUnlock()
	return len(fake.findByNameArgsForCall)
}

func (fake *FakeOrganizationRepository) FindByNameArgsForCall(i int) string {
	fake.findByNameMutex.RLock()
	defer fake.findByNameMutex.RUnlock()
	return fake.findByNameArgsForCall[i].name
}

func (fake *FakeOrganizationRepository) FindByNameReturns(result1 models.Organization, result2 error) {
	fake.FindByNameStub = nil
	fake.findByNameReturns = struct {
		result1 models.Organization
		result2 error
	}{result1, result2}
}

func (fake *FakeOrganizationRepository) Create(org models.Organization) (apiErr error) {
	fake.createMutex.Lock()
	fake.createArgsForCall = append(fake.createArgsForCall, struct {
		org models.Organization
	}{org})
	fake.recordInvocation("Create", []interface{}{org})
	fake.createMutex.Unlock()
	if fake.CreateStub != nil {
		return fake.CreateStub(org)
	} else {
		return fake.createReturns.result1
	}
}

func (fake *FakeOrganizationRepository) CreateCallCount() int {
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	return len(fake.createArgsForCall)
}

func (fake *FakeOrganizationRepository) CreateArgsForCall(i int) models.Organization {
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	return fake.createArgsForCall[i].org
}

func (fake *FakeOrganizationRepository) CreateReturns(result1 error) {
	fake.CreateStub = nil
	fake.createReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeOrganizationRepository) Rename(orgGUID string, name string) (apiErr error) {
	fake.renameMutex.Lock()
	fake.renameArgsForCall = append(fake.renameArgsForCall, struct {
		orgGUID string
		name    string
	}{orgGUID, name})
	fake.recordInvocation("Rename", []interface{}{orgGUID, name})
	fake.renameMutex.Unlock()
	if fake.RenameStub != nil {
		return fake.RenameStub(orgGUID, name)
	} else {
		return fake.renameReturns.result1
	}
}

func (fake *FakeOrganizationRepository) RenameCallCount() int {
	fake.renameMutex.RLock()
	defer fake.renameMutex.RUnlock()
	return len(fake.renameArgsForCall)
}

func (fake *FakeOrganizationRepository) RenameArgsForCall(i int) (string, string) {
	fake.renameMutex.RLock()
	defer fake.renameMutex.RUnlock()
	return fake.renameArgsForCall[i].orgGUID, fake.renameArgsForCall[i].name
}

func (fake *FakeOrganizationRepository) RenameReturns(result1 error) {
	fake.RenameStub = nil
	fake.renameReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeOrganizationRepository) Delete(orgGUID string) (apiErr error) {
	fake.deleteMutex.Lock()
	fake.deleteArgsForCall = append(fake.deleteArgsForCall, struct {
		orgGUID string
	}{orgGUID})
	fake.recordInvocation("Delete", []interface{}{orgGUID})
	fake.deleteMutex.Unlock()
	if fake.DeleteStub != nil {
		return fake.DeleteStub(orgGUID)
	} else {
		return fake.deleteReturns.result1
	}
}

func (fake *FakeOrganizationRepository) DeleteCallCount() int {
	fake.deleteMutex.RLock()
	defer fake.deleteMutex.RUnlock()
	return len(fake.deleteArgsForCall)
}

func (fake *FakeOrganizationRepository) DeleteArgsForCall(i int) string {
	fake.deleteMutex.RLock()
	defer fake.deleteMutex.RUnlock()
	return fake.deleteArgsForCall[i].orgGUID
}

func (fake *FakeOrganizationRepository) DeleteReturns(result1 error) {
	fake.DeleteStub = nil
	fake.deleteReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeOrganizationRepository) SharePrivateDomain(orgGUID string, domainGUID string) (apiErr error) {
	fake.sharePrivateDomainMutex.Lock()
	fake.sharePrivateDomainArgsForCall = append(fake.sharePrivateDomainArgsForCall, struct {
		orgGUID    string
		domainGUID string
	}{orgGUID, domainGUID})
	fake.recordInvocation("SharePrivateDomain", []interface{}{orgGUID, domainGUID})
	fake.sharePrivateDomainMutex.Unlock()
	if fake.SharePrivateDomainStub != nil {
		return fake.SharePrivateDomainStub(orgGUID, domainGUID)
	} else {
		return fake.sharePrivateDomainReturns.result1
	}
}

func (fake *FakeOrganizationRepository) SharePrivateDomainCallCount() int {
	fake.sharePrivateDomainMutex.RLock()
	defer fake.sharePrivateDomainMutex.RUnlock()
	return len(fake.sharePrivateDomainArgsForCall)
}

func (fake *FakeOrganizationRepository) SharePrivateDomainArgsForCall(i int) (string, string) {
	fake.sharePrivateDomainMutex.RLock()
	defer fake.sharePrivateDomainMutex.RUnlock()
	return fake.sharePrivateDomainArgsForCall[i].orgGUID, fake.sharePrivateDomainArgsForCall[i].domainGUID
}

func (fake *FakeOrganizationRepository) SharePrivateDomainReturns(result1 error) {
	fake.SharePrivateDomainStub = nil
	fake.sharePrivateDomainReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeOrganizationRepository) UnsharePrivateDomain(orgGUID string, domainGUID string) (apiErr error) {
	fake.unsharePrivateDomainMutex.Lock()
	fake.unsharePrivateDomainArgsForCall = append(fake.unsharePrivateDomainArgsForCall, struct {
		orgGUID    string
		domainGUID string
	}{orgGUID, domainGUID})
	fake.recordInvocation("UnsharePrivateDomain", []interface{}{orgGUID, domainGUID})
	fake.unsharePrivateDomainMutex.Unlock()
	if fake.UnsharePrivateDomainStub != nil {
		return fake.UnsharePrivateDomainStub(orgGUID, domainGUID)
	} else {
		return fake.unsharePrivateDomainReturns.result1
	}
}

func (fake *FakeOrganizationRepository) UnsharePrivateDomainCallCount() int {
	fake.unsharePrivateDomainMutex.RLock()
	defer fake.unsharePrivateDomainMutex.RUnlock()
	return len(fake.unsharePrivateDomainArgsForCall)
}

func (fake *FakeOrganizationRepository) UnsharePrivateDomainArgsForCall(i int) (string, string) {
	fake.unsharePrivateDomainMutex.RLock()
	defer fake.unsharePrivateDomainMutex.RUnlock()
	return fake.unsharePrivateDomainArgsForCall[i].orgGUID, fake.unsharePrivateDomainArgsForCall[i].domainGUID
}

func (fake *FakeOrganizationRepository) UnsharePrivateDomainReturns(result1 error) {
	fake.UnsharePrivateDomainStub = nil
	fake.unsharePrivateDomainReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeOrganizationRepository) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.listOrgsMutex.RLock()
	defer fake.listOrgsMutex.RUnlock()
	fake.getManyOrgsByGUIDMutex.RLock()
	defer fake.getManyOrgsByGUIDMutex.RUnlock()
	fake.findByNameMutex.RLock()
	defer fake.findByNameMutex.RUnlock()
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	fake.renameMutex.RLock()
	defer fake.renameMutex.RUnlock()
	fake.deleteMutex.RLock()
	defer fake.deleteMutex.RUnlock()
	fake.sharePrivateDomainMutex.RLock()
	defer fake.sharePrivateDomainMutex.RUnlock()
	fake.unsharePrivateDomainMutex.RLock()
	defer fake.unsharePrivateDomainMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeOrganizationRepository) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ organizations.OrganizationRepository = new(FakeOrganizationRepository)
