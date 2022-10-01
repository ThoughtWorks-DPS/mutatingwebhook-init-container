#!/bin/bash

docker pull twdps/sidecar-mutatingwebhook-init-container
export RESULT=$(docker run -it twdps/sidecar-mutatingwebhook-init-container:latest | grep "Success: created mutatingwebhookconfiguration")
if [[ "${RESULT}" == "" ]]; then
  echo 'Container did not log success'
  exit 1
fi
