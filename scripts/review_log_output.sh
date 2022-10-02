#!/bin/bash

docker pull twdps/mutatingwebhook-init-container
export RESULT=$(docker run -it twdps/mutatingwebhook-init-container:latest  | grep "Requested MutatingWebhookConfiguration")
if [[ "${RESULT}" == "" ]]; then
  echo 'Container did not log success'
  exit 1
fi
