#!/bin/bash

cf uninstall-plugin copy
go build
cf install-plugin -f cf-copy-plugin
rm cf-copy-plugin