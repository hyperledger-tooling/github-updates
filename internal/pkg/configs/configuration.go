/**
 * Copyright 2021 Hyperledger Community
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package configs

import (
	"github-updates/internal/pkg/utils"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// Configuration reads the input config file
type Configuration struct {
	GlobalConfiguration GlobalConfiguration      `yaml:"global"`
	Issues              IssueConfiguration       `yaml:"issues"`
	PullRequests        PullRequestConfiguration `yaml:"pull-requests"`
	Releases            ReleaseConfiguration     `yaml:"releases"`
}

type IssueConfiguration struct {
	IssueTags               []string                `yaml:"issue-tags"`
	IssueCreatedHistoryDays int                     `yaml:"created-history-days"`
	IssueSummaryFileName    string                  `yaml:"summary-filename"`
	IssueReportShouldRun    bool                    `yaml:"should-run"`
	IssueDataFile           string                  `yaml:"data-file"`
	IssueExternalTemplate   ElementExternalTemplate `yaml:"external-template"`
}

type PullRequestConfiguration struct {
	PRSummaryFileName  string                  `yaml:"summary-filename"`
	PRReportShouldRun  bool                    `yaml:"should-run"`
	PRDataFile         string                  `yaml:"data-file"`
	PRExternalTemplate ElementExternalTemplate `yaml:"external-template"`
}

type ReleaseConfiguration struct {
	ReleaseSummaryFileName  string                  `yaml:"summary-filename"`
	ReleaseReportShouldRun  bool                    `yaml:"should-run"`
	ReleaseDataFile         string                  `yaml:"data-file"`
	ReleaseExternalTemplate ElementExternalTemplate `yaml:"external-template"`
}

type ElementExternalTemplate struct {
	Input     string `yaml:"input"`
	Output    string `yaml:"output"`
	Summary   string `yaml:"summary"`
	Generated string `yaml:"sum-generated"`
}

type GlobalConfiguration struct {
	Organizations    []Organization   `yaml:"organizations"`
	DaysCount        int              `yaml:"scrape-duration-days"`
	ExternalTemplate ExternalTemplate `yaml:"external-template"`
}

type ExternalTemplate struct {
	Enabled     bool   `yaml:"enabled"`
	TemplateFor string `yaml:"template-for"`
}

// Organization represents GitHub organization
type Organization struct {
	Organization OrganizationStructure `yaml:"organization"`
}

// OrganizationStructure has information about particular
// GitHub Organization
type OrganizationStructure struct {
	Github string `yaml:"github"`
	Name   string `yaml:"name"`
}

// ReadConfiguration returns the configuration object
func ReadConfiguration() Configuration {

	log.Println("Reading the configuration file")
	var config Configuration
	var configFile = utils.GetEnvOrDefault(ConfigFile, "config.yaml")
	fileContents, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Couldn't read the config file %v, Err: %v", configFile, err)
	}
	err = yaml.Unmarshal(fileContents, &config)
	if err != nil {
		log.Fatalf("Error while parsing the config file %v, Err: %v", configFile, err)
	}
	return config
}
