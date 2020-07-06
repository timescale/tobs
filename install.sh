#!/bin/sh

set -eu

INSTALLROOT=${INSTALLROOT:-"${HOME}/.ts-obs"}
TS_OBS_VERSION=${TS_OBS_VERSION:-0.1.0-alpha.4.1}

happyexit() {
  echo ""
  echo "Add the ts-obs CLI to your path with:"
  echo ""
  echo "  export PATH=\$PATH:${INSTALLROOT}/bin"
  echo ""
  echo "After starting your cluster, run"
  echo ""
  echo "  ts-obs install"
  echo ""
  exit 0
}

validate_checksum() {
  filename=$1
  SHA=$(curl -sfL "${url}.sha256")
  echo ""
  echo "Validating checksum..."

  case $checksumbin in
    *openssl)
      checksum=$($checksumbin dgst -sha256 "${filename}" | sed -e 's/^.* //')
      ;;
    *shasum)
      checksum=$($checksumbin -a256 "${filename}" | sed -e 's/^.* //')
      ;;
  esac

  if [ "$checksum" != "$SHA" ]; then
    echo "Checksum validation failed." >&2
    return 1
  fi
  echo "Checksum valid."
  return 0
}

OS=$(uname -s)
arch=$(uname -m)
case $OS in
  CYGWIN* | MINGW64*)
    OS=windows.exe
    ;;
  Darwin)
    ;;
  Linux)
    case $arch in
      x86_64)
        ;;
      *)
        echo "Timescale Observability does not support $OS/$arch. Please open an issue with your platform details."
        exit 1
        ;;
    esac
    ;;
  *)
    echo "Timescale Observability does not support $OS/$arch. Please open an issue with your platform details."
    exit 1
    ;;
esac
OS=$(echo $OS | tr '[:upper:]' '[:lower:]')

checksumbin=$(command -v openssl) || checksumbin=$(command -v shasum) || {
  echo "Failed to find checksum binary. Please install openssl or shasum."
  exit 1
}

tmpdir=$(mktemp -d /tmp/ts-obs.XXXXXX)
srcfile="ts-obs-${TS_OBS_VERSION}-${OS}"
dstfile="${INSTALLROOT}/bin/ts-obs-${TS_OBS_VERSION}"
url="https://github.com/timescale/timescale-observability/releases/download/${TS_OBS_VERSION}/${srcfile}"

if [ -e "${dstfile}" ]; then
  if validate_checksum "${dstfile}"; then
    echo ""
    echo "ts-obs ${TS_OBS_VERSION} was already downloaded; making it the default"
    echo ""
    echo "To force re-downloading, delete '${dstfile}' then run me again."
    (
      rm -f "${INSTALLROOT}/bin/ts-obs"
      ln -s "${dstfile}" "${INSTALLROOT}/bin/ts-obs"
    )
    happyexit
  fi
fi

(
  cd "$tmpdir"

  echo "Downloading ${srcfile}..."
  curl -fLO "${url}"
  echo "Download complete!"

  if ! validate_checksum "${srcfile}"; then
    exit 1
  fi
  echo ""
)

(
  mkdir -p "${INSTALLROOT}/bin"
  mv "${tmpdir}/${srcfile}" "${dstfile}"
  chmod +x "${dstfile}"
  rm -f "${INSTALLROOT}/bin/ts-obs"
  ln -s "${dstfile}" "${INSTALLROOT}/bin/ts-obs"
)

rm -r "$tmpdir"
echo "ts-obs ${TS_OBS_VERSION} was successfully installed 🎉"
echo ""
happyexit
