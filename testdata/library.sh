#!/usr/bin/env bash

#?/name        string-utils
#?/version     1.0.0
#?/description
 # A library of string manipulation functions.
 ##

#@/public
 # Converts a string to uppercase.
 #
 # @operand <string>    The string to convert
 # @stdout              Uppercase result
 ##
to_upper() {
    echo "${1^^}"
}

#@/private
 # Internal helper for validation.
 ##
_validate_input() {
    [[ -n "$1" ]]
}
