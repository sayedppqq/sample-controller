#!/bin/bash

set -x

vendor/k8s.io/code-generator/generate-groups.sh all \
  github.com/sayedppqq/sample-controller/pkg/client \
  github.com/sayedppqq/sample-controller/pkg/apis \
  sayedppqq.dev:v1alpha1 \
  --go-header-file /home/appscodepc/go/src/github.com/sayedppqq/sample-controller/hack/boilerplate.go.txt