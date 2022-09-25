#!/usr/bin/env bats

@test "init complete" {
  run bash -c "kubectl get po -n ci-dev"
  [[ "${output}" =~ "Running" ]]
}

@test "review init container logs" {
  run bash -c 'kubectl logs $(kubectl get pod -l app=tls-test-app -n ci-dev -o jsonpath="{.items[0].metadata.name}") -n ci-dev -c sidecar-mutatingwebhook-init-container'
  [[ "${output}" =~ "Success:" ]]
}

@test "review mytatingwebhookconfiguration deployments" {
  run bash -c 'kubectl get mutatingwebhookconfigurations'
  [[ "${output}" =~ "opa-injector-admission-controller-webhook" ]]
}

@test "confirm certs written to emptyDir" {
  run bash -c 'kubectl exec -it $(kubectl get pod -l app=tls-test-app -n ci-dev -o jsonpath="{.items[0].metadata.name}") -n ci-dev -- cat /etc/tls/tls.crt | grep CERTIFI
CATE'
  [[ "${output}" =~ "CERTIFICATE" ]]
}
