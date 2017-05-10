package command

import (
	"os"
	"strings"

	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"

	"code.cloudfoundry.org/cli/plugin"
)

// CopyPlugin -
type CopyPlugin struct {
	ui      terminal.UI
	copyCmd CopyCmd
}

// NewCopyPlugin -
func NewCopyPlugin(copyCmd CopyCmd) *CopyPlugin {
	return &CopyPlugin{
		copyCmd: copyCmd,
	}
}

// Start -
func (c *CopyPlugin) Start() {
	plugin.Start(c)
}

// GetMetadata -
func (c *CopyPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "copy",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 0,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "copy",
				HelpText: "Copy current space artifacts to another space. Uses targets saved by 'Targets' plugin when copying to another Cloud Foundry target.",
				UsageDetails: plugin.Usage{
					Usage: "cf copy DEST_SPACE [DEST_ORG] [DEST_TARGET] " +
						"[--apps|-a APPLICATIONS] [--host-format|-n HOST_FORMAT] [--domain|-d DOMAIN] [--droplet] " +
						"[--ups|-s COPY_AS_UPS] [--services-only|-o] [--recreate-services|-r]" +
						"[-debug|-d]",
					Options: map[string]string{
						"-apps, -a":              "Copy only the given applications and their bound services. Default is to copy all applications.",
						"-host-format, -n":       "Format of app route's hostname to make it unique i.e. \"{{.host}}-{{.space}}\".",
						"-domain, -m":            "Domain to use to create routes for copied apps with same hostname.",
						"-droplet, -c":           "Application droplet will be copied to the destination as is. Otherwise, the application bits will be re-pushed.",
						"-ups, -s":               "Comma separated list of service instances that will be copied as user provided services in the target space.",
						"-service-types, -t":     "Comma separated list of service types that will be copied as user provided services in the target space.",
						"-services-only, -o":     "Make copies of services only. If a list of applications are provided then only services bound to that app will be copied.",
						"-recreate-services, -r": "Recreates services at destination.",
						"-debug, -d":             "Output debug messages.",
					},
				},
			},
		},
	}
}

// Run -
func (c *CopyPlugin) Run(cliConnection plugin.CliConnection, args []string) {

	c.ui = terminal.NewUI(os.Stdin, os.Stdout, terminal.NewTeePrinter(os.Stdout), trace.NewLogger(os.Stdout, false, "", ""))

	switch args[0] {
	case "copy":
		if o, ok := c.parseCopyOptions(args[1:]); ok {
			c.copyCmd.Execute(cliConnection, o)
		}
	default:
		return
	}
}

func (c *CopyPlugin) parseCopyOptions(args []string) (*CopyOptions, bool) {

	var i int

	o := CopyOptions{}

	for i, arg := range args {
		if strings.Index(arg, "-") == 0 {
			break
		}
		switch i {
		case 0:
			o.DestSpace = arg
		case 1:
			o.DestOrg = arg
		case 2:
			o.DestTarget = arg
		default:
			c.ui.Failed("Invalid positional argument '%s'.", arg)
			return nil, false
		}
	}

	if o.DestSpace == "" {
		c.ui.Failed("At least a destination space must be provided.")
		return nil, false
	}

	f := flags.New()
	f.NewStringFlag("apps", "a", "")
	f.NewStringFlag("host-format", "n", "")
	f.NewStringFlag("domain", "m", "")
	f.NewBoolFlag("droplet", "c", "")
	f.NewStringFlag("ups", "s", "")
	f.NewStringFlag("service-types", "t", "")
	f.NewBoolFlag("services-only", "o", "")
	f.NewBoolFlag("recreate-services", "r", "")
	f.NewBoolFlag("debug", "d", "")

	err := f.Parse(args[i:]...)
	if err != nil {
		c.ui.Failed(err.Error())
		return nil, false
	}
	if f.IsSet("apps") {
		o.SourceAppNames = strings.Split(f.String("apps"), ",")
	}
	if f.IsSet("host-format") {
		o.AppHostFormat = f.String("host-format")
	}
	if f.IsSet("domain") {
		o.AppRouteDomain = f.String("domain")
	}
	if f.IsSet("droplet") {
		o.CopyAsDroplet = f.Bool("droplet")
	}
	if f.IsSet("ups") {
		o.ServiceInstancesToCopyAsUPS = strings.Split(f.String("ups"), ",")
	}
	if f.IsSet("service-types") {
		o.ServiceTypesToCopyAsUPS = strings.Split(f.String("service-types"), ",")
	}
	if f.IsSet("recreate-services") {
		o.RecreateServices = f.Bool("recreate-services")
	}
	if f.IsSet("services-only") {
		o.ServicesOnly = f.Bool("services-only")
	}
	if f.IsSet("debug") {
		o.Debug = f.Bool("debug")
	}
	trace := os.Getenv("CF_TRACE")
	if trace != "" {
		o.Debug = true
		o.TracePath = trace
	}
	return &o, true
}
