#!/usr/bin/env bash

##
 # Installs all dependencies listed in the specified package definition (i.e.
 # package.sh).
 #
 # @arg package_definition  the package definition file to be parsed
 # @arg target_directory    the target directory for the dependenices
 #
 # @exit 0 on success
 #
deps() {
    local _root="$( cd $( dirname "$( [ -z "${BASH_SOURCE[0]}" ] && echo "$0" || echo "${BASH_SOURCE[0]}" )" )/.. && pwd $@ )"
    local _package="${1:-"$_root/package.sh"}"
    local _deps="${2:-"$_root/deps"}"

    source "$_package"
    [ -d "$_deps" ] && rm -rf $_deps
    mkdir -p "$_deps"
    cd "$_deps"

    for dep in "${dependencies[@]}"; do
        git clone "$dep"
    done
}

# Only execute if this file isn't being sourced.
if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    deps "$@"
fi
