#!/usr/bin/env bash

# This script has regular comments but no shedoc documentation.

greet() {
    local name="${1:-World}"
    echo "Hello, $name!"
}

greet "$@"
