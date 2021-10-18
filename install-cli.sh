#!/bin/sh

set -eu

INSTALLROOT=${INSTALLROOT:-"${HOME}/.tobs"}
TOBS_VERSION=${TOBS_VERSION:-0.7.0}

happyexit() {
	cat <<-EOF
		Add the tobs CLI to your system binaries with:

		  sudo cp ${INSTALLROOT}/tobs /usr/local/bin

		Alternatively, add tobs to your path in the current session with: export PATH=\$PATH:${INSTALLROOT}

		After starting your Kubernetes cluster, run

		  tobs install

	EOF
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

install_tobs() {
	OS=$(uname -s)
	arch=$(uname -m)
	case $OS in
	Darwin) ;;

	Linux)
		case $arch in
		x86_64) ;;

		i386) ;;

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
	dstfile="${INSTALLROOT}/tobs-${TOBS_VERSION}"
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
		mkdir -p "${INSTALLROOT}"
		mv "${tmpdir}/${srcfile}" "${dstfile}"
		chmod +x "${dstfile}"
		rm -f "${INSTALLROOT}/tobs"
		ln -s "${dstfile}" "${INSTALLROOT}/tobs"

		if [ ! -L "/usr/local/bin/tobs" ]; then
			echo "Attempting to link ${INSTALLROOT}/tobs to /usr/local/bin to easier binary discovery."
			# Following command shouldn't stop installation when sudo prompt is canceled
			sudo ln -s "${INSTALLROOT}/tobs" "/usr/local/bin/tobs" || :
		fi
	)

	rm -r "$tmpdir"
	echo "tobs ${TOBS_VERSION} was successfully installed 🎉"
	echo ""
	happyexit
}
install_tobs
