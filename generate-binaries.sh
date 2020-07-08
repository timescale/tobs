#!/bin/sh

GOPATH=${GOPATH:-"${HOME}/go"}
TS_OBS_VERSION=${TS_OBS_VERSION:-0.1.0-alpha.4.1}

(
  cd ts-obs

  export GOOS=linux
  export GOARCH=amd64
  go install

  export GOOS=darwin
  export GOARCH=amd64
  go install

  export GOOS=windows
  export GOARCH=amd64
  go install
)

(
  cd ${GOPATH}/bin

  mv ts-obs ts-obs-${TS_OBS_VERSION}-linux
  openssl dgst -sha256 "ts-obs-${TS_OBS_VERSION}-linux" | sed -e 's/^.* //' > "ts-obs-${TS_OBS_VERSION}-linux.sha256"

  cd ${GOPATH}/bin/darwin_amd64

  mv ts-obs ts-obs-${TS_OBS_VERSION}-darwin
  openssl dgst -sha256 "ts-obs-${TS_OBS_VERSION}-darwin" | sed -e 's/^.* //' > "ts-obs-${TS_OBS_VERSION}-darwin.sha256"

  cd ${GOPATH}/bin/windows_amd64

  mv ts-obs.exe ts-obs-${TS_OBS_VERSION}-windows.exe
  openssl dgst -sha256 "ts-obs-${TS_OBS_VERSION}-windows.exe" | sed -e 's/^.* //' > "ts-obs-${TS_OBS_VERSION}-windows.exe.sha256"
)
