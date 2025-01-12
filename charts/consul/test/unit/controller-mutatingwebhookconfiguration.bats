#!/usr/bin/env bats

load _helpers

@test "controller/MutatingWebhookConfiguration: enabled by default" {
  cd `chart_dir`
  local actual=$(helm template \
      -s templates/controller-mutatingwebhookconfiguration.yaml  \
      . | tee /dev/stderr |
      yq 'length > 0' | tee /dev/stderr)
  [ "${actual}" = "true" ]
}

@test "controller/MutatingWebhookConfiguration: enabled with controller.enabled=true" {
  cd `chart_dir`
  local actual=$(helm template \
      -s templates/controller-mutatingwebhookconfiguration.yaml  \
      --set 'controller.enabled=true' \
      . | tee /dev/stderr |
      yq 'length > 0' | tee /dev/stderr)
  [ "${actual}" = "true" ]
}
