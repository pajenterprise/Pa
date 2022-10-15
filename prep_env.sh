#!/bin/bash

export SRC_PATH=/go/src/github.com/DataDog/datadog-agent
export OMNIBUS_BASE_DIR="/.omnibus"
mkdir -p $OMNIBUS_BASE_DIR
export RELEASE_VERSION_6=nightly
export RELEASE_VERSION_7=nightly-a7
export USE_S3_CACHING=""
export CONDA_ENV=ddpy3
export AGENT_MAJOR_VERSION=6
export PYTHON_RUNTIMES="2,3"
export PACKAGE_ARCH=amd64
export DESTINATION_DEB="datadog-agent_6_amd64.deb"
export DESTINATION_DBG_DEB="datadog-agent-dbg_6_amd64.deb"
export RELEASE_VERSION=$RELEASE_VERSION_6
export OMNIBUS_RUBY_VERSION="jaime/hackadog"
export OMNIBUS_SOFTWARE_VERSION="jaime/hackadog"

source /root/.bashrc
conda activate $CONDA_ENV
