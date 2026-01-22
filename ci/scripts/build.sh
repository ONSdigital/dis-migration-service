#!/bin/bash -eux

# Build the application
pushd pull_request
  make build
  cp build/dis-migration-service Dockerfile.concourse ../build
popd
