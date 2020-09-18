#!/usr/bin/env bash

##
 # Installs all dependencies listed in the specified package definition (i.e.
 # package.sh).
 #
 # @arg package_definition  the package definition file to be parsed
 # @arg target_directory    the target directory for the dependenices
 #
 # @exit 0 on success
 # @exit 1 when a package definition isn't found
 # @exit 2 when installation fails
 #
deps() {
    local E_PACKAGE_DEF=1
    local E_DEPS_INSALL=2

    local _root="$(
        while [ ! -e "package.sh" ]; do
            if [[ $PWD != / ]]; then
                cd ..
            else
                return $E_PACKAGE_DEPS
            fi
        done

        echo "$PWD"
    )"

    local _package="${1:-"$_root/package.sh"}"
    local _deps="${2:-"$_root/deps"}"

    source "$_package" || return $E_PACKAGE_DEF

    (
        [ -d "$_deps" ] && rm -rf $_deps
        mkdir -p "$_deps" || return $E_DEPS_INSALL
        cd "$_deps"

        for dep in "${dependencies[@]}"; do
            git clone "$dep" || return $E_DEPS_INSTALL
        done
    )
}

# Only execute if this file isn't being sourced.
if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then deps "$@"; fi
