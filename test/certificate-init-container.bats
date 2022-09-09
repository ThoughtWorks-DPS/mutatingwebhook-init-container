#!/usr/bin/env bats

@test "init complete" {
  run bash -c "kubectl get po -n ci-dev"
  [[ "${output}" =~ "Running" ]]
}

@test "review init container logs" {
  run bash -c 'kubectl logs $(kubectl get pod -l app=tls-test-app -n ci-dev -o jsonpath="{.items[0].metadata.name}") -n ci-dev -c certificate-init-container'
  [[ "${output}" =~ "commonName: init-container-test" ]]
}

@test "review tls-test-app logs" {
  run bash -c 'kubectl logs $(kubectl get pod -l app=tls-test-app -n ci-dev -o jsonpath="{.items[0].metadata.name}") -n ci-dev -c tls-test-app'
  [[ "${output}" =~ "Loading TLS certificates" ]]
}

@test "confirm certs written to emptyDir" {
  run bash -c 'kubectl exec -it $(kubectl get pod -l app=tls-test-app -n ci-dev -o jsonpath="{.items[0].metadata.name}") -n ci-dev -- ls -la /etc/tls'
  [[ "${output}" =~ "tls.crt" ]]
}
