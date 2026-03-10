#!/bin/sh
set -e

# First-run: neither config nor workspace exists.
# If config.json is already mounted but workspace is missing we skip onboard to
# avoid the interactive "Overwrite? (y/n)" prompt hanging in a non-TTY container.
if [ ! -d "${HOME}/.fernwood/workspace" ] && [ ! -f "${HOME}/.fernwood/config.json" ]; then
    fernwood onboard
    echo ""
    echo "First-run setup complete."
    echo "Edit ${HOME}/.fernwood/config.json (add your API key, etc.) then restart the container."
    exit 0
fi

exec fernwood gateway "$@"
