#!/bin/bash

version=$($(dirname "${BASH_SOURCE}")/print-version.sh)

echo "-ldflags \"-X github.com/darkowlzz/clouddev/version.Version=${version}\""
