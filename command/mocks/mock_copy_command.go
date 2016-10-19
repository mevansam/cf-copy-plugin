package mock_test

import (
	"os"

	"github.com/mevansam/cf-copy-plugin/command"

	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/plugin"
)

// MockCopyCommand -
type MockCopyCommand struct {
	validate func(o *command.CopyOptions)
}

// NewMockCopyCommand -
func NewMockCopyCommand(validate func(o *command.CopyOptions)) *MockCopyCommand {
	return &MockCopyCommand{validate}
}

// Execute -
func (m MockCopyCommand) Execute(cli plugin.CliConnection, o *command.CopyOptions) {
	m.validate(o)

	logger := trace.NewLogger(os.Stdout, true, "", "")
	ui := terminal.NewUI(os.Stdin, os.Stdout, terminal.NewTeePrinter(os.Stdout), logger)
	ui.Say("Done")
}
