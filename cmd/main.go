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

package main

import (
	client2 "hyperledger-updates/internal/pkg/client"
	"hyperledger-updates/internal/pkg/configs"
	"hyperledger-updates/internal/pkg/utils"
	"log"
	"os"
	"path"
	"path/filepath"
)

var AppVersion = ""

const AppName = "Hyperledger Updates"

func init() {
	if AppVersion == "" {
		AppVersion = "Unknown" // expect to set the version at build time
	}
}

func main() {

	log.Printf("%s version: %s\n", AppName, AppVersion)
	config := configs.ReadConfiguration()
	client := client2.NewClient()
	log.Println("Listing repositories for each organization")

	expectedPrList, orgReleasesList, issueList, errorOccurred :=
		getExpectedReportsLists(config, client)
	if errorOccurred {
		return
	}
	externalPRList, externalReleaseList, externalIssueList :=
		getExternalReports(config, expectedPrList, orgReleasesList, issueList)
	var reportFilePath, templateFilePath string
	var err error

	if config.PullRequests.PRReportShouldRun {
		// Save noteworthy PRs into a file
		reportFilePath =
			utils.GetEnvOrDefault(
				configs.PrSummaryFilePath,
				config.PullRequests.PRSummaryFileName,
			)
		templateFilePath = utils.GetEnvOrDefault(configs.PRTemplateFile, "html/template/pr-template.html")
		err =
			generateReport(
				config.PullRequests.PRDataFile,
				expectedPrList,
				reportFilePath,
				templateFilePath,
			)
		if err != nil {
			log.Fatalf("Failed to generate the report: %v, with template: %v. Error is: %v", reportFilePath, templateFilePath, err)
		}
		err =
			generateExternalPR(
				config.PullRequests.PRExternalTemplate,
				externalPRList,
			)
		if err != nil {
			log.Fatalf("Failed to generate the report: %v, with template: %v. Error is: %v",
				config.PullRequests.PRExternalTemplate.Output, config.PullRequests.PRExternalTemplate.Input, err)
		}
	}

	if config.Releases.ReleaseReportShouldRun {
		// Save releases into a file
		reportFilePath =
			utils.GetEnvOrDefault(
				configs.ReleaseSummaryFilePath,
				config.Releases.ReleaseSummaryFileName,
			)
		templateFilePath = utils.GetEnvOrDefault(configs.ReleaseTemplateFile, "html/template/release-template.html")
		err =
			generateReport(
				config.Releases.ReleaseDataFile,
				orgReleasesList,
				reportFilePath,
				templateFilePath,
			)
		if err != nil {
			log.Fatalf("Err: %v", err)
		}
		err =
			generateExternalRelease(
				config.Releases.ReleaseExternalTemplate,
				externalReleaseList,
			)
		if err != nil {
			log.Fatalf("Failed to generate the report: %v, with template: %v. Error is: %v",
				config.Releases.ReleaseExternalTemplate.Output, config.Releases.ReleaseExternalTemplate.Input, err)
		}
	}

	if config.Issues.IssueReportShouldRun {
		// Save releases into a file
		reportFilePath =
			utils.GetEnvOrDefault(
				configs.IssueSummaryFilePath,
				config.Issues.IssueSummaryFileName,
			)
		templateFilePath = utils.GetEnvOrDefault(configs.IssueTemplateFile, "html/template/issue-template.html")
		err =
			generateReport(
				config.Issues.IssueDataFile,
				issueList,
				reportFilePath,
				templateFilePath,
			)
		if err != nil {
			log.Fatalf("Err: %v", err)
		}
		err =
			generateExternalIssue(
				config.Issues.IssueExternalTemplate,
				externalIssueList,
			)
		if err != nil {
			log.Fatalf("Failed to generate the report: %v, with template: %v. Error is: %v",
				config.Issues.IssueExternalTemplate.Output, config.Issues.IssueExternalTemplate.Input, err)
		}
	}
}

func getOrg(
	organizations []configs.Organization,
	name string,
) configs.OrganizationStructure {
	for _, org := range organizations {
		if org.Organization.Github == name {
			return org.Organization
		}
	}
	// Unexpected
	log.Fatalln("Unexpected organization found!")
	return configs.OrganizationStructure{}
}

func getExternalReports(config configs.Configuration,
	expectedPrList []configs.PullRequestDetails,
	orgReleasesList []configs.ReleaseDetails,
	issueList []configs.IssueDetails,
) ([]configs.ExternalPRDetails, []configs.ExternalReleaseDetails, []configs.ExternalIssueDetails) {
	if !config.GlobalConfiguration.ExternalTemplate.Enabled {
		return []configs.ExternalPRDetails{},
			[]configs.ExternalReleaseDetails{},
			[]configs.ExternalIssueDetails{}
	}
	var externalPRDetails []configs.ExternalPRDetails
	for _, org := range expectedPrList {
		organization := getOrg(
			config.GlobalConfiguration.Organizations,
			org.Organization,
		)
		for _, repo := range org.PrRepoLists {
			elementPRDetails := configs.ExternalPRDetails{
				Organization: configs.OrganizationStructure{
					Github: org.Organization,
					Name:   organization.Name,
				},
				Repository: configs.RepositoryStructure{
					Name: repo.Repository,
					Link: "https://github.com/" + org.Organization + "/" + repo.Repository,
				},
				PRs: repo.PRs,
			}
			externalPRDetails = append(externalPRDetails, elementPRDetails)
		}
	}
	var externalReleaseDetails []configs.ExternalReleaseDetails
	for _, org := range orgReleasesList {
		organization := getOrg(
			config.GlobalConfiguration.Organizations,
			org.Organization,
		)
		for _, repo := range org.ReleaseRepoLists {
			elementRelease := configs.ExternalReleaseDetails{
				Organization: configs.OrganizationStructure{
					Github: org.Organization,
					Name:   organization.Name,
				},
				Repository: configs.RepositoryStructure{
					Name: repo.Repository,
					Link: "https://github.com/" + org.Organization + "/" + repo.Repository,
				},
				Releases: repo.Releases,
			}
			externalReleaseDetails = append(externalReleaseDetails, elementRelease)
		}
	}
	var externalIssueDetails []configs.ExternalIssueDetails
	for _, org := range issueList {
		organization := getOrg(
			config.GlobalConfiguration.Organizations,
			org.Organization,
		)
		for _, repo := range org.IssueLists {
			elementIssue := configs.ExternalIssueDetails{
				Organization: configs.OrganizationStructure{
					Github: org.Organization,
					Name:   organization.Name,
				},
				Repository: configs.RepositoryStructure{
					Name: repo.Repository,
					Link: "https://github.com/" + org.Organization + "/" + repo.Repository,
				},
				Issues: repo.Issues,
			}
			externalIssueDetails = append(externalIssueDetails, elementIssue)
		}
	}
	return externalPRDetails, externalReleaseDetails, externalIssueDetails
}

func generateExternalPR(
	externalTemplate configs.ElementExternalTemplate,
	values []configs.ExternalPRDetails,
) error {
	if len(values) == 0 {
		log.Println("External template file generation is not requested")
		return nil
	}
	for _, value := range values {
		err :=
			generateExternalFile(
				value,
				value.Repository.Name,
				value.Organization.Github,
				externalTemplate,
			)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateExternalIssue(
	externalTemplate configs.ElementExternalTemplate,
	values []configs.ExternalIssueDetails,
) error {
	if len(values) == 0 {
		log.Println("External template file generation is not requested")
		return nil
	}
	for _, value := range values {
		err :=
			generateExternalFile(
				value,
				value.Repository.Name,
				value.Organization.Github,
				externalTemplate,
			)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateExternalRelease(
	externalTemplate configs.ElementExternalTemplate,
	values []configs.ExternalReleaseDetails,
) error {
	if len(values) == 0 {
		log.Println("External template file generation is not requested")
		return nil
	}
	for _, value := range values {
		err :=
			generateExternalFile(
				value,
				value.Repository.Name,
				value.Organization.Github,
				externalTemplate,
			)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateExternalFile(
	value interface{},
	filename string,
	org string,
	externalTemplate configs.ElementExternalTemplate,
) error {
	var err error
	outputFileName := filename + filepath.Ext(externalTemplate.Input)
	outputPath := path.Join(externalTemplate.Output, org)
	err = os.MkdirAll(outputPath, 755)
	if err != nil {
		return err
	}
	outputFilePath := path.Join(outputPath, outputFileName)
	err = utils.PrettyPrint(value, outputFilePath, externalTemplate.Input)
	if err != nil {
		return err
	}
	return nil
}

func generateReport(
	dataFileName string,
	v interface{},
	reportFilePath string,
	templateFilePath string,
) error {

	err := utils.SaveIntoFile(v, dataFileName)
	if err != nil {
		log.Fatalf("Error in saving report as json : %v. Error is: %v", reportFilePath, err)
		return err
	}

	err = utils.PrettyPrint(v, reportFilePath, templateFilePath)
	if err != nil {
		log.Fatalf("Error in generating report html: %v. Error is: %v", reportFilePath, err)
		return err
	}

	return nil
}

func getExpectedReportsLists(
	config configs.Configuration,
	client client2.GHClientInterface,
) ([]configs.PullRequestDetails, []configs.ReleaseDetails, []configs.IssueDetails, bool) {
	var expectedPrList []configs.PullRequestDetails
	var orgReleasesList []configs.ReleaseDetails
	var issueList []configs.IssueDetails

	for _, organization := range config.GlobalConfiguration.Organizations {

		repos, err := client.ListRepositories(organization.Organization.Github)
		if err != nil {
			log.Fatalf("Err: %v", err)
			return nil, nil, nil, true
		}
		log.Printf("List for %v is : %v", organization, repos)

		if config.PullRequests.PRReportShouldRun {
			//// Pull requests
			expectedPrs, errorOccurred :=
				getExpectedPullRequests(client, organization, repos, config)
			if errorOccurred {
				return nil, nil, nil, errorOccurred
			}
			expectedPrList = append(expectedPrList, expectedPrs)
		}

		if config.Releases.ReleaseReportShouldRun {
			// Releases
			releaseList, errorOccurred :=
				getReleaseList(client, organization, repos, config)
			if errorOccurred {
				return nil, nil, nil, errorOccurred
			}
			orgReleasesList = append(orgReleasesList, releaseList)
		}

		if config.Issues.IssueReportShouldRun {
			//good first issues and other configured tags
			expectedIssues, errorOccurred :=
				getIssueList(client, organization, repos, config)
			if errorOccurred {
				return nil, nil, nil, errorOccurred
			}
			issueList = append(issueList, expectedIssues)
		}
	}
	return expectedPrList, orgReleasesList, issueList, false
}

func getReleaseList(
	client client2.GHClientInterface,
	organization configs.Organization,
	repos []string,
	config configs.Configuration,
) (configs.ReleaseDetails, bool) {
	orgReleases, err :=
		client.ListReleases(
			organization.Organization.Github,
			repos,
			config.GlobalConfiguration.DaysCount,
		)
	if err != nil {
		log.Fatalf("Err: %v", err)
		return configs.ReleaseDetails{}, true
	}
	releaseList := configs.ReleaseDetails{
		Organization:     organization.Organization.Github,
		ReleaseRepoLists: orgReleases,
	}
	return releaseList, false
}

func getIssueList(
	client client2.GHClientInterface,
	organization configs.Organization,
	repos []string,
	config configs.Configuration,
) (configs.IssueDetails, bool) {
	issues, err :=
		client.IssueWithLabels(
			organization.Organization.Github,
			repos,
			config.Issues.IssueTags,
			config.Issues.IssueCreatedHistoryDays,
		)
	if err != nil {
		log.Fatalf("Err: %v", err)
		return configs.IssueDetails{}, true
	}
	issueList := configs.IssueDetails{
		Organization: organization.Organization.Github,
		IssueLists:   issues,
	}
	return issueList, false
}

func getExpectedPullRequests(
	client client2.GHClientInterface,
	organization configs.Organization,
	repos []string,
	config configs.Configuration,
) (configs.PullRequestDetails, bool) {
	pRs, err :=
		client.ListPRs(
			organization.Organization.Github,
			repos,
			config.GlobalConfiguration.DaysCount,
		)
	if err != nil {
		log.Fatalf("Err: %v", err)
		return configs.PullRequestDetails{}, true
	}
	expectedPrs := configs.PullRequestDetails{
		Organization: organization.Organization.Github,
		PrRepoLists:  pRs,
	}
	return expectedPrs, false
}
