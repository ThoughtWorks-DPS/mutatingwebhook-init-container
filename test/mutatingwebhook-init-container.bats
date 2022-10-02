#!/usr/bin/env bats

@test "init containers completed successfully" {
  run bash -c "kubectl get po -n ci-dev"
  [[ "${output}" =~ "Running" ]]
}

@test "confirm certificate is written to emptyDir" {
  run bash -c 'kubectl exec -it $(kubectl get pod -l app=tls-test-app -n ci-dev -o jsonpath="{.items[0].metadata.name}") -n ci-dev -- cat /etc/tls/tls.crt | grep CERTIFICATE'
  [[ "${output}" =~ "CERTIFICATE" ]]
}

@test "review mutatingwebhook-init-container logs for Success message" {
  run bash -c 'kubectl logs $(kubectl get pod -l app=tls-test-app -n ci-dev -o jsonpath="{.items[0].metadata.name}") -n ci-dev -c mutatingwebhook-init-container'
  [[ "${output}" =~ "Success:" ]]
}

@test "review deployed mutatingwebhookconfigurations for presence of test hook" {
  run bash -c 'kubectl get mutatingwebhookconfigurations'
  [[ "${output}" =~ "opa-injector-admission-controller-webhook" ]]
}
