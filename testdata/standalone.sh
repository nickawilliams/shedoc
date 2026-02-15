#!/usr/bin/env bash

#?/name    greet
#?/version 1.0.0

#@/command
 # Prints a greeting message.
 #
 # @operand [name=World]              Name to greet
 # @exit    0                         Success
 # @stdout                            Greeting message
 ##

echo "Hello, ${1:-World}!"
