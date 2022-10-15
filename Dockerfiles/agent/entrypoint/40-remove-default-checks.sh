#!/bin/bash

if [[ -z "${DD_REMOVE_DEFAULT_CHECKS}" ]]; then
    exit 0
fi

# Remove all default checks
find /etc/datadog-agent/conf.d/ -iname "*.yaml.default" -delete
