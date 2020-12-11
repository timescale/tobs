#!/bin/sh

set -eu

INSTALLROOT=${INSTALLROOT:-"${HOME}/.tobs"}
TOBS_VERSION=${TOBS_VERSION:-0.1.3}

happyexit() {
  echo ""
  echo "Add the tobs CLI to your system binaries with:"
  echo ""
  echo "  sudo cp ${INSTALLROOT}/bin/tobs /usr/local/bin"
  echo ""
  echo "Alternatively, add tobs to your path in the current session with: export PATH=\$PATH:${INSTALLROOT}/bin"
  echo ""
  echo "After starting your Kubernetes cluster, run"
  echo ""
  echo "  tobs install"
  echo ""
  exit 0
}

validate_checksum() {
  filename=$1
  checksumlist=$(curl -sfL "${url}/checksums.txt")
  echo ""
  echo "Validating checksum..."

  checksum=$($checksumbin -a256 "${filename}")

  if echo "${checksumlist}" | grep -Fxq "${checksum}"; then
    echo "Checksum valid."
    return 0
  else
    echo "Checksum validation failed." >&2
    return 1
  fi
}

install_tobs () {
  OS=$(uname -s)
  arch=$(uname -m)
  case $OS in
    Darwin)
      ;;
    Linux)
      case $arch in
        x86_64)
          ;;
        i386)
          ;;
        *)
          echo "The Observability Stack does not support $OS/$arch. Please open an issue with your platform details."
          exit 1
          ;;
      esac
      ;;
    *)
      echo "The Observability Stack does not support $OS/$arch. Please open an issue with your platform details."
      exit 1
      ;;
  esac

  checksumbin=$(command -v shasum) || {
    echo "Failed to find checksum binary. Please install shasum."
    exit 1
  }

  tmpdir=$(mktemp -d /tmp/tobs.XXXXXX)
  srcfile="tobs_${TOBS_VERSION}_${OS}_${arch}"
  dstfile="${INSTALLROOT}/bin/tobs-${TOBS_VERSION}"
  url="https://github.com/timescale/tobs/releases/download/${TOBS_VERSION}"

  (
    cd "$tmpdir"

    echo "\n"
    echo "Downloading ${srcfile}..."
    curl --proto '=https' --tlsv1.2 -sSfLO "${url}/${srcfile}"
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
    rm -f "${INSTALLROOT}/bin/tobs"
    ln -s "${dstfile}" "${INSTALLROOT}/bin/tobs"
  )

  rm -r "$tmpdir"
  echo "tobs ${TOBS_VERSION} was successfully installed 🎉"
  echo ""
  happyexit
}
install_tobs
