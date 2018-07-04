#!/bin/bash

if ! [ -x "$(which go-bindata)" ]; then
  echo ''
  echo 'Error: go-bindata is not installed.' >&2
  echo 'Run "go get -u github.com/jteeuwen/go-bindata/..."' >&2

  echo '(the "..." are important!)'
  echo ''
  echo 'Do not forget to add $GOPATH/bin to your $PATH :)'
  echo ''
  exit 1
fi

go-bindata -pkg assetsManager -o assets_manager/manager.go assets
