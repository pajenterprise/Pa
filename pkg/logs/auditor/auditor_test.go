// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package auditor

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/pkg/logs/config"
	"github.com/DataDog/datadog-agent/pkg/logs/message"
	"github.com/stretchr/testify/suite"
)

var testpath = "testpath"

type AuditorTestSuite struct {
	suite.Suite
	testDir  string
	testPath string
	testFile *os.File

	inputChan chan message.Message
	a         *Auditor
	source    *config.LogSource
}

func (suite *AuditorTestSuite) SetupTest() {
	suite.testDir = "tests/"
	os.Remove(suite.testDir)
	os.MkdirAll(suite.testDir, os.ModeDir)
	suite.testPath = fmt.Sprintf("%s/auditor.json", suite.testDir)

	_, err := os.Create(suite.testPath)
	suite.Nil(err)

	suite.inputChan = make(chan message.Message)
	suite.a = New(suite.inputChan, "")
	suite.a.registryPath = suite.testPath
	suite.source = config.NewLogSource("", &config.LogsConfig{Path: testpath})
}

func (suite *AuditorTestSuite) TearDownTest() {
	os.Remove(suite.testDir)
}

func (suite *AuditorTestSuite) TestAuditorUpdatesRegistry() {
	suite.a.registry = make(map[string]*RegistryEntry)
	suite.Equal(0, len(suite.a.registry))
	suite.a.updateRegistry(suite.source.Config.Path, "42")
	suite.Equal(1, len(suite.a.registry))
	suite.Equal("42", suite.a.registry[suite.source.Config.Path].Offset)
	suite.a.updateRegistry(suite.source.Config.Path, "43")
	suite.Equal(1, len(suite.a.registry))
	suite.Equal("43", suite.a.registry[suite.source.Config.Path].Offset)
}

func (suite *AuditorTestSuite) TestAuditorFlushesAndRecoversRegistry() {
	suite.a.registry = make(map[string]*RegistryEntry)
	suite.a.registry[suite.source.Config.Path] = &RegistryEntry{
		LastUpdated: time.Date(2006, time.January, 12, 1, 1, 1, 1, time.UTC),
		Offset:      "42",
	}
	suite.a.flushRegistry()
	r, err := ioutil.ReadFile(suite.testPath)
	suite.Nil(err)
	suite.Equal("{\"Version\":2,\"Registry\":{\"testpath\":{\"LastUpdated\":\"2006-01-12T01:01:01.000000001Z\",\"Offset\":\"42\"}}}", string(r))

	suite.a.registry = make(map[string]*RegistryEntry)
	suite.a.registry = suite.a.recoverRegistry()
	suite.Equal("42", suite.a.registry[suite.source.Config.Path].Offset)
}

func (suite *AuditorTestSuite) TestAuditorRecoversRegistryForOffset() {
	suite.a.registry = make(map[string]*RegistryEntry)
	suite.a.registry[suite.source.Config.Path] = &RegistryEntry{
		Offset: "42",
	}

	offset := suite.a.GetLastCommittedOffset(suite.source.Config.Path)
	suite.Equal("42", offset)

	othersource := config.NewLogSource("", &config.LogsConfig{Path: "anotherpath"})
	offset = suite.a.GetLastCommittedOffset(othersource.Config.Path)
	suite.Equal("", offset)
}

func (suite *AuditorTestSuite) TestAuditorCleansupRegistry() {
	suite.a.registry = make(map[string]*RegistryEntry)
	suite.a.registry[suite.source.Config.Path] = &RegistryEntry{
		LastUpdated: time.Date(2006, time.January, 12, 1, 1, 1, 1, time.UTC),
		Offset:      "42",
	}

	otherpath := "otherpath"
	suite.a.registry[otherpath] = &RegistryEntry{
		LastUpdated: time.Now().UTC(),
		Offset:      "43",
	}
	suite.a.flushRegistry()
	suite.Equal(2, len(suite.a.registry))

	suite.a.cleanupRegistry()
	suite.Equal(1, len(suite.a.registry))
	suite.Equal("43", suite.a.registry[otherpath].Offset)
}

func TestScannerTestSuite(t *testing.T) {
	suite.Run(t, new(AuditorTestSuite))
}
