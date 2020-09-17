PATH := deps/bin:$(PATH)		# Adds dependency executables to PATH
SHELL := /bin/bash 				# macOS requires this to export PATH

# Installs dependencies.
# ----------------------------------------------------------------------------
deps:
	./scripts/deps.sh

# Cleans the project of all generated files.
# ----------------------------------------------------------------------------
clean:
	cat .cleanrc | sed -E '/^#.*$$/ d' | sed '/^\\s*$$/ d' | xargs rm -rf

# Runs all tests.
# ----------------------------------------------------------------------------
test:

.PHONY: clean deps test