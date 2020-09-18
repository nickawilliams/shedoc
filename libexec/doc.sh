#!/usr/bin/env bash

#?/synopsis     doc file [tag]
#?/summary      parses shedoc comments in the specified file and outputs it

#?/description
#? Inspects the file for shedoc comments, and outputs the requested tag (or
#? description by default).

doc() {
	local file="$1"
	local tag="description"
	local result=""

	if [[ $# -gt 1 ]]; then
		tag=$(printf "%s" "$2" | sed -E -e 's/[\/&]/\\&/g')
	fi

	result=$(
		# Grab the file contents & extract target comment block.
		sed -n -E -e "/\#\?\/${tag}\h*/,/^(\#\?\/.*){0,1}$/p" "$file" |

		# Delete last line (since the above is inclusive on the end pattern).
		sed '$d' |

		# Remove the starting line if it only contains the tag.
		sed -E -e "/\#\?\/${tag}\w*$/d" |

		# Remove the tag syntax from the starting line if followed by text.
		sed -E -e "s/\#\?(\/${tag}[	 ]*|[ ])//" |

		# Remove the comment syntax if it is a continuation comment.
		sed -E -e "s/\#\?//" |

		# Remove leading space padding.
		sed 's/^[[:space:]]//'
	)

	if [[ ! -z "$result" ]]; then
		echo "${result//
/\\n}"
		return 0
	fi

	return 1
}

# Only execute if this file isn't being sourced.
if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then deps "$@"; fi
