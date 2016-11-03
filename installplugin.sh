#!/bin/bash

cf uninstall-plugin CopyPlugin
go build
cf install-plugin -f cf-copy-plugin
rm cf-copy-plugin