#!/bin/sh
set -e
# Sense installer for Linux/MacOS script
#
# See https://github.com/qlik-oss/sense-installer for the installation steps.
#
# This script is meant for quick & easy install via:
#   $ curl -fsSL https://raw.githubusercontent.com/qlik-oss/sense-installer/scripts/install.sh | sh -

RELEASES_URI="https://api.github.com/repos/qlik-oss/sense-installer/releases/latest"
REPO_URL="https://github.com/qlik-oss/sense-installer"
BIN_DIR="/usr/local/bin"
FILE_NAME="kubectl-qliksense"

command_exists() {
	command -v "$@" > /dev/null 2>&1
}

do_install() {
	echo "\n==> Executing qliksense install script\n"

	if ! command_exists kubectl; then
		echo "\n==> ERROR: kubectl is required for $FILE_NAME to work"
		echo "See https://github.com/qlik-oss/sense-installer#requirements for more information\n"
		exit 1
	fi

	if ! command_exists curl; then
		echo "==> ERROR: curl is missing"
		exit 1
	fi

	user="$(id -un 2>/dev/null || true)"

	SUDO='sh -c'
	if [ "$(id -u)" != "0" ]; then
		root_msg="\tNext operation might ask for root password to place\n\t$FILE_NAME in $BIN_DIR and set executable permission\n"
		if command_exists sudo; then
			echo $root_msg
			SUDO='sudo -E sh -c'
		elif command_exists su; then
			echo $root_msg
			SUDO='su -c'
		else
			cat >&2 <<-'EOF'
			Error: this installer needs the ability to run commands as root.
			We are unable to find either "sudo" or "su" available to make this happen.
			EOF
			exit 1
		fi
	fi

	if command_exists curl; then
		releases=$(mktemp)

		if [ -n "$GITHUB_TOKEN" ]; then
			curl -v -H "Authorization: token $GITHUB_TOKEN" $RELEASES_URI > $releases 2>&1
		else
			curl -v $RELEASES_URI > $releases 2>&1
		fi

		if ! grep -q "Status: 200" $releases; then
			echo "==> ERROR: cannot get qliksense-installer"
			echo "GitHub:" $(grep "message" $releases | cut -d '"' -f 4) "\n"
			echo "Use: GITHUB_TOKEN=token install-sense-cli.sh"
			exit 1
		fi

		download_url=$(grep -E "browser_download_url.*linux" $releases | grep -v "tar.gz" | cut -d '"' -f 4)

		echo "==> Installing to $BIN_DIR/$FILE_NAME\n"
		if [ -n "$GITHUB_TOKEN" ]; then
			$SUDO "curl -fSL -H \"Authorization: token $GITHUB_TOKEN\" $download_url -o $BIN_DIR/$FILE_NAME"
		else
			$SUDO "curl -fSL $download_url -o $BIN_DIR/$FILE_NAME"
		fi
		$SUDO "chmod +x $BIN_DIR/$FILE_NAME"
	fi

	if command_exists $FILE_NAME; then
		echo "\n==> Success: You can now start using $FILE_NAME\n"
	else
		echo "\n==> ERROR: Something went wrong, try again"
	fi
}

do_install
