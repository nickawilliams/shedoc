#!/bin/bash

#?/name edge-cases

#@/
 # Bare visibility defaults to public.
 ##
bare_func() {
    echo "bare"
}

#@/public
 # A function declared with the function keyword.
 ##
function keyword_func {
    echo "keyword"
}
