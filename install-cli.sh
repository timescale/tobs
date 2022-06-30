#!/bin/sh

set -eu

INSTALLROOT=${INSTALLROOT:-"${HOME}/.local/bin"}
TOBS_VERSION=${TOBS_VERSION:-0.11.3}

happyexit() {
	local symlink_msg=""
	if [ "$1" = "1" ]; then
		symlink_msg=" and for your coinvenience linked to /usr/local/bin/tobs"
	fi
	cat <<-EOF

		tobs ${TOBS_VERSION} was successfully installed 🎉

		Binary is available at ${INSTALLROOT}/tobs${symlink_msg}.

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
	)

	rm -r "$tmpdir"
	echo "tobs ${TOBS_VERSION} was successfully installed 🎉"
	echo ""

	local symlink=0
	if [ ! -L "/usr/local/bin/tobs" ]; then
		echo "Attempting to link ${INSTALLROOT}/tobs to /usr/local/bin for easier binary discovery."
		# Following command shouldn't stop installation when sudo prompt is canceled
		if timeout --foreground 120s sudo ln -s "${INSTALLROOT}/tobs" "/usr/local/bin/tobs" 2> /dev/null; then
			symlink=1
		else
			echo "Proceeding without creating symlink at /usr/local/bin/tobs"
		fi
	fi

	happyexit $symlink
}
install_tobs
