Cloud Foundry CLI Copy Plugin
===

This plugin allows you to copy artifacts from one Cloud Foundry space to another space within the same org or 
different org within the same deployment. It can also copy artifacts to a different Cloud Foundry target that 
has been saved using the '[CF Targets](https://github.com/guidowb/cf-targets-plugin)' plugin. Use the plugin 
options to selectively copy just the service instances only or everything including applications.

# Usage
```
$ cf copy --help
NAME:
   copy - Copy current space artifacts to another space. Uses targets saved by 'Targets' plugin when copying to another Cloud Foundry target.

USAGE:
   cf copy DEST_SPACE [DEST_ORG] [DEST_TARGET] [--apps|-a APPLICATIONS] [--host-format|-n HOST_FORMAT] [--domain|-d DOMAIN] [--droplet] [--ups|-s COPY_AS_UPS] [--services-only|-o] [--recreate-services|-r][-debug|-d]

OPTIONS:
   --apps, -a                    Copy only the given applications and their bound services. Default is to copy all applications.
   --host-format, -n             Format of app route's hostname to make it unique i.e. "{{.host}}-{{.space}}".
   --domain, -m                  Domain to use to create routes for copied apps with same hostname.
   --droplet, -c                 Application droplet will be copied to the destination as is. Otherwise, the application bits will be re-pushed.
   --ups, -s                     Comma separated list of services that will be copied as user provided services in the target space.
   --recreate-services, -r       Recreates services at destination.
   --services-only, -o           Make copies of services only. If a list of applications are provided then only services bound to that app will be copied.
   --debug, -d                   Output debug messages.
```

# Installation

## Install from CLI
```
$ cf add-plugin-repo CF-Community http://plugins.cloudfoundry.org/
$ cf install-plugin 'copy' -r CF-Community
```

## Install from Source (need to have [Go](http://golang.org/dl/) installed)
```
$ go get github.com/cloudfoundry/cli
$ go get github.com/mevansam/cf-copy-plugin
$ cd $GOPATH/src/github.com/mevansam/cf-copy-plugin
$ go build
$ cf install-plugin -f cf-copy-plugin
```
