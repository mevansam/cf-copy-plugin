package command_test

import (
	. "code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "code.cloudfoundry.org/cli/utils/testhelpers/io"
	. "github.com/mevansam/cf-copy-plugin/command"
	. "github.com/mevansam/cf-copy-plugin/command/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Copy Plugin Tests", func() {
	var (
		fakeCliConnection *FakeCliConnection
		copyPlugin        *CopyPlugin
	)

	BeforeEach(func() {
		fakeCliConnection = &FakeCliConnection{}
		copyPlugin = &CopyPlugin{}
	})

	Describe("GetMetadata", func() {

		It("Returns metadata", func() {
			md := copyPlugin.GetMetadata()
			Expect(md).NotTo(BeNil())
		})

		It("Has a help message", func() {
			md := copyPlugin.GetMetadata()
			Expect(md.Commands[0].HelpText).NotTo(BeNil())
		})
	})

	Describe("Run Copy Command", func() {

		It("Should parse all args", func() {

			copyPluginFake := NewCopyPlugin(NewMockCopyCommand(func(o *CopyOptions) {
				Expect(o.DestSpace).To(Equal("fake_space"))
				Expect(o.DestOrg).To(Equal("fake_org"))
				Expect(o.DestTarget).To(Equal("fake_target"))
				Expect(o.SourceAppNames[0]).To(Equal("fake_app1"))
				Expect(o.SourceAppNames[1]).To(Equal("fake_app2"))
				Expect(o.AppHostFormat).To(Equal("fake_host_format"))
				Expect(o.AppRouteDomain).To(Equal("fake_domain"))
				Expect(o.CopyAsDroplet).To(BeTrue())
				Expect(o.CopyAsUpsServices[0]).To(Equal("fake_svc1"))
				Expect(o.CopyAsUpsServices[1]).To(Equal("fake_svc2"))
				Expect(o.RecreateServices).To(BeTrue())
				Expect(o.ServicesOnly).To(BeTrue())
			}))

			output := CaptureOutput(func() {
				copyPluginFake.Run(fakeCliConnection, []string{
					"copy",
					"fake_space",
					"fake_org",
					"fake_target",
					"--apps", "fake_app1,fake_app2",
					"--host-format", "fake_host_format",
					"--domain", "fake_domain",
					"--droplet",
					"--ups", "fake_svc1,fake_svc2",
					"--recreate-services",
					"--services-only",
				})
			})

			Expect(output[0]).To(Equal("Done"))
		})

		It("Should accept minimal args", func() {

			copyPluginFake := NewCopyPlugin(NewMockCopyCommand(func(o *CopyOptions) {
				Expect(o.DestSpace).To(Equal("fake_space"))
				Expect(o.DestOrg).To(Equal(""))
				Expect(o.DestTarget).To(Equal(""))
				Expect(o.SourceAppNames).To(BeEmpty())
				Expect(o.CopyAsUpsServices).To(BeEmpty())
				Expect(o.ServicesOnly).To(BeFalse())
			}))

			output := CaptureOutput(func() {
				copyPluginFake.Run(fakeCliConnection, []string{
					"copy",
					"fake_space",
				})
			})

			Expect(output[0]).To(Equal("Done"))
		})

		It("Should accept fewer but valid args", func() {

			copyPluginFake := NewCopyPlugin(NewMockCopyCommand(func(o *CopyOptions) {
				Expect(o.DestSpace).To(Equal("fake_space"))
				Expect(o.DestOrg).To(Equal("fake_org"))
				Expect(o.DestTarget).To(Equal(""))
				Expect(o.SourceAppNames[0]).To(Equal("fake_app"))
				Expect(o.CopyAsUpsServices).To(BeEmpty())
				Expect(o.ServicesOnly).To(BeFalse())
			}))

			output := CaptureOutput(func() {
				copyPluginFake.Run(fakeCliConnection, []string{
					"copy",
					"fake_space",
					"fake_org",
					"--apps", "fake_app",
				})
			})

			Expect(output[0]).To(Equal("Done"))
		})

		It("Should recognize missing space", func() {

			copyPluginFake := NewCopyPlugin(NewMockCopyCommand(func(o *CopyOptions) {
				Fail("CLI argument parsing should have failed and been handled.")
			}))

			output := CaptureOutput(func() {
				copyPluginFake.Run(fakeCliConnection, []string{
					"copy",
					"--apps", "fake_app",
					"--ups", "fake_svc1,fake_svc2",
					"--services-only",
				})
			})

			Expect(output[0]).To(Equal("FAILED"))
			Expect(output[1]).To(Equal("At least a destination space must be provided."))
		})

		It("Should not accept extra positional arg", func() {

			copyPluginFake := NewCopyPlugin(NewMockCopyCommand(func(o *CopyOptions) {
				Fail("CLI argument parsing should have failed and been handled.")
			}))

			output := CaptureOutput(func() {
				copyPluginFake.Run(fakeCliConnection, []string{
					"copy",
					"fake_space",
					"fake_org",
					"fake_target",
					"invalid_argument",
				})
			})

			Expect(output[0]).To(Equal("FAILED"))
			Expect(output[1]).To(Equal("Invalid positional argument 'invalid_argument'."))
		})
	})
})
