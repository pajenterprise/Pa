// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.

// +build windows
// +build python

package app

import (
	"fmt"
	"path/filepath"

	"github.com/DataDog/datadog-agent/pkg/util/winutil"
)

const (
	pythonBin = "python.exe"
)

func getRelPyPath() string {
	return filepath.Join(fmt.Sprintf("embedded%s", pythonMajorVersion), pythonBin)
}

func getRelChecksPath() (string, error) {
	return filepath.Join(fmt.Sprintf("embedded%s", pythonMajorVersion), "Lib", "site-packages", "datadog_checks"), nil
}

func validateUser(allowRoot bool) error {
	elevated, _ := winutil.IsProcessElevated()
	if !elevated {
		return fmt.Errorf("Operation is not possible for unelevated process.")
	}
	return nil
}
