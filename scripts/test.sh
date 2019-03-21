#!/bin/bash
# This script performs tests against the dcos-metrics project, specifically:
#
#   * gofmt         (https://golang.org/cmd/gofmt)
#   * goimports     (https://godoc.org/cmd/goimports)
#   * golint        (https://github.com/golang/lint)
#   * go vet        (https://golang.org/cmd/vet)
#   * test coverage (https://blog.golang.org/cover)
#   * misspell      (https://github.com/client9/misspell)
#
# It outputs test and coverage reports in a way that Jenkins can understand,
# with test results in JUnit format and test coverage in Cobertura format.
# The reports are saved to build/$SUBDIR/{test-reports,coverage-reports}/*.xml
#
set -e
set -o pipefail
export PATH="${GOPATH}/bin:${PATH}"

PACKAGES="$(go list ./... | grep -v /vendor/)"
FILES="$(find ./ -name "*.go" | grep -v vendor)"
SOURCE_DIR=$(git rev-parse --show-toplevel)
BUILD_DIR="${SOURCE_DIR}/build"


function logmsg {
    echo -e "\n\n*** $1 ***\n"
}

function _gofmt {
    logmsg "Running 'gofmt' ..."
    test -z "$(gofmt -l ${FILES} | tee /dev/stderr)"
}

function _misspell {
    logmsg "Running 'misspell' ..."
    GO111MODULE=off go get -u github.com/client9/misspell/cmd/misspell
    misspell -error ${FILES}
}

function _goimports {
    logmsg "Running 'goimports' ..."
    GO111MODULE=off go get -u golang.org/x/tools/cmd/goimports
    test -z "$(gofmt -l -d ${FILES} | tee /dev/stderr)"
}

function _golint {
    logmsg "Running 'go lint' ..."
    GO111MODULE=off go get -u golang.org/x/lint/golint
    for pkg in $PACKAGES; do
        golint -set_exit_status $pkg
    done
}

function _govet {
    logmsg "Running 'go vet' ..."
    go vet ${PACKAGES}
}

function _unittest_with_coverage {
	go test -cover -race -test.v ${PACKAGES}
}


# Main.
function main {
    _goimports
    _misspell
    _gofmt
    _golint
    _govet
    _unittest_with_coverage
}

main
