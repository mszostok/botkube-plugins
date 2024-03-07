#!/usr/bin/env sh

#exit on error 
set -o errexit
set -o pipefail

echo 'level=info msg="Analytics disabled via configuration settings."'
echo 'level=info msg="Registering filter "ObjectAnnotationChecker" (enabled: true)...  component="Filter Engine"'
echo 'level=info msg="Registering filter "NodeEventsChecker" (enabled: true)...  component="Filter Engine"'
echo 'level=info msg="Starting Backup Job                                  integration=psql"'
echo 'level=info msg="Allocating 200M of memory for Backup Job             integration=psql"'
stress --vm 1 --vm-bytes 250M --vm-hang 1 --timeout 5s --verbose
echo 'level=info msg="Backup completed...'
