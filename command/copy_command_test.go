package command_test

import (
	"strings"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/organizations"
	"code.cloudfoundry.org/cli/cf/api/spaces"
	"code.cloudfoundry.org/cli/cf/models"
	. "code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "code.cloudfoundry.org/cli/utils/testhelpers/io"
	. "github.com/mevansam/cf-copy-plugin/command"
	. "github.com/mevansam/cf-copy-plugin/command/mocks"
	. "github.com/mevansam/cf-copy-plugin/copy/mocks"
	"github.com/mevansam/cf-copy-plugin/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Copy Command Tests", func() {

	var (
		fakeCliConnection       *FakeCliConnection
		mockTargets             *MockTargets
		mockSessionProvider     *MockSessionProvider
		mockSrcSession          *MockSession
		mockDestSession         *MockSession
		mockApplicationsManager *MockApplicationsManager
		mockServicesManager     *MockServicesManager
		copyCommand             CopyCmd
	)

	BeforeEach(func() {
		fakeCliConnection = &FakeCliConnection{}

		mockTargets = &MockTargets{
			CurrentTarget: "fake_source_target",
			Targets: map[string]string{
				"fake_source_target": "/fake/source/target.json",
				"fake_dest_target":   "/fake/dest/target.json",
			},
		}

		mockSrcSession = &MockSession{}
		mockDestSession = &MockSession{}

		mockSessionProvider = &MockSessionProvider{make(map[string]helpers.CloudControllerSession)}
		mockSessionProvider.MockSessionMap["/fake/source/target.json"] = mockSrcSession
		mockSessionProvider.MockSessionMap["/fake/dest/target.json"] = mockDestSession

		mockApplicationsManager = &MockApplicationsManager{}
		mockServicesManager = &MockServicesManager{}
		copyCommand = NewCopyCommand(mockTargets, mockSessionProvider, mockApplicationsManager, mockServicesManager)
	})

	Context("Test initialization", func() {

		It("Recognizes cf-targets plugin has not been installed", func() {
			fakeCliConnection.CliCommandWithoutTerminalOutputStub = func(args ...string) ([]string, error) {
				return strings.Split(cf_plugins_out_1, "\n"), nil
			}
			output := CaptureOutput(func() {
				copyCommand.Execute(fakeCliConnection, &CopyOptions{})
			})
			Expect(output[0]).To(Equal("FAILED"))
			Expect(output[1]).To(Equal("'Targets' plugin is requried to determine destination Cloud Foundry target."))
		})
		It("Recognizes that the given target does not exist", func() {
			fakeCliConnection.CliCommandWithoutTerminalOutputStub = func(args ...string) ([]string, error) {
				return strings.Split(cf_plugins_out_2, "\n"), nil
			}
			output := CaptureOutput(func() {
				copyCommand.Execute(fakeCliConnection, &CopyOptions{
					DestSpace:      "fake_dest_space",
					DestOrg:        "fake_dest_org",
					DestTarget:     "fake_unknown_dest_target",
					SourceAppNames: []string{"fake_source_app"},
				})
			})
			Expect(output[0]).To(Equal("FAILED"))
			Expect(output[1]).To(Equal("A target named 'fake_unknown_dest_target' cannot be found."))
		})
		It("Recognizes that the current session does not have a target", func() {
			fakeCliConnection.CliCommandWithoutTerminalOutputStub = func(args ...string) ([]string, error) {
				return strings.Split(cf_plugins_out_2, "\n"), nil
			}
			mockSrcSession.MockHasTarget = func() bool { return false }
			output := CaptureOutput(func() {
				copyCommand.Execute(fakeCliConnection, &CopyOptions{
					DestSpace:      "fake_dest_space",
					DestOrg:        "fake_dest_org",
					DestTarget:     "fake_dest_target",
					SourceAppNames: []string{"fake_source_app"},
				})
			})
			Expect(output[0]).To(Equal("FAILED"))
			Expect(output[1]).To(Equal("The CLI target org and space needs to be set."))
		})
		It("Recognizes that the source and destination are the same", func() {
			fakeCliConnection.CliCommandWithoutTerminalOutputStub = func(args ...string) ([]string, error) {
				return strings.Split(cf_plugins_out_2, "\n"), nil
			}
			mockSrcSession.MockHasTarget = func() bool { return true }
			mockSrcSession.MockGetSessionOrg = func() models.OrganizationFields { return models.OrganizationFields{Name: "fake_dest_org"} }
			mockSrcSession.MockGetSessionSpace = func() models.SpaceFields { return models.SpaceFields{Name: "fake_dest_space"} }
			output := CaptureOutput(func() {
				copyCommand.Execute(fakeCliConnection, &CopyOptions{
					DestSpace:      "fake_dest_space",
					DestOrg:        "fake_dest_org",
					SourceAppNames: []string{"fake_source_app"},
				})
			})
			Expect(output[0]).To(Equal("FAILED"))
			Expect(output[1]).To(Equal("The source and destination are the same."))
		})
	})

	Context("Test copy", func() {

		BeforeEach(func() {
			fakeCliConnection.CliCommandWithoutTerminalOutputStub = func(args ...string) ([]string, error) {
				return strings.Split(cf_plugins_out_2, "\n"), nil
			}
			mockSrcSession.MockHasTarget = func() bool { return true }
			mockSrcSession.MockGetSessionOrg = func() models.OrganizationFields { return models.OrganizationFields{Name: "fake_src_org"} }
			mockSrcSession.MockGetSessionSpace = func() models.SpaceFields { return models.SpaceFields{Name: "fake_src_space"} }
			mockSrcSession.MockAppSummary = func() api.AppSummaryRepository {
				return &FakeAppSummaryRepository{
					GetSummariesInCurrentSpaceStub: func() (apps []models.Application, err error) {
						apps = []models.Application{models.Application{}}
						apps[0].Name = "fake_source_app"
						return
					},
				}
			}
			mockDestSession.MockGetSessionUsername = func() string { return "fake_user" }
			mockDestSession.MockOrganizations = func() organizations.OrganizationRepository {
				return &FakeOrganizationRepository{
					FindByNameStub: func(name string) (org models.Organization, apiErr error) {
						Expect(name).To(Equal("fake_dest_org"))
						org = models.Organization{}
						org.GUID = "1234"
						org.Name = name
						return
					},
				}
			}
			mockDestSession.MockSpaces = func() spaces.SpaceRepository {
				return &FakeSpaceRepository{
					FindByNameInOrgStub: func(name, orgGUID string) (space models.Space, apiErr error) {
						Expect(name).To(Equal("fake_dest_space"))
						Expect(orgGUID).To(Equal("1234"))
						space = models.Space{}
						space.Name = name
						return
					},
				}
			}
			mockDestSession.MockSetSessionOrg = func(org models.OrganizationFields) {
				Expect(org.Name).To(Equal("fake_dest_org"))
			}
			mockDestSession.MockSetSessionSpace = func(space models.SpaceFields) {
				Expect(space.Name).To(Equal("fake_dest_space"))
			}
		})

		It("Should set the target org and space", func() {
			output := CaptureOutput(func() {
				copyCommand.Execute(fakeCliConnection, &CopyOptions{
					DestSpace:      "fake_dest_space",
					DestOrg:        "fake_dest_org",
					DestTarget:     "fake_dest_target",
					SourceAppNames: []string{"fake_source_app"},
				})
			})
			Expect(output[2]).To(Equal("OK"))
		})
	})
})

const cf_plugins_out_1 = `Listing Installed Plugins...
OK

Plugin Name        Version   Command Name       Command Help
pcfdev             0.19.0    dev, pcfdev        Control PCF Dev VMs running on your workstation
FirehosePlugin     0.11.0    nozzle             Displays messages from the firehose
FirehosePlugin     0.11.0    app-nozzle         Displays messages from the firehose for a given app
`

const cf_plugins_out_2 = `Listing Installed Plugins...
OK

Plugin Name        Version   Command Name       Command Help
pcfdev             0.19.0    dev, pcfdev        Control PCF Dev VMs running on your workstation
FirehosePlugin     0.11.0    nozzle             Displays messages from the firehose
FirehosePlugin     0.11.0    app-nozzle         Displays messages from the firehose for a given app
cf-targets         1.1.0     targets            List available targets
cf-targets         1.1.0     set-target         Set current target
cf-targets         1.1.0     save-target        Save current target
cf-targets         1.1.0     delete-target      Delete a saved target
`
