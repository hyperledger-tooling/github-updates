package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	yaml "gopkg.in/yaml.v2"
)

const (
	// ConfigFile env variable that can be overridden
	ConfigFile = "CONFIG_FILE"
	// GitHubToken env variable
	GitHubToken = "GITHUB_TOKEN"
	// PrSummaryFilePath env variable
	PrSummaryFilePath = "PR_SUMMARY_FILE_PATH"
	// ReleaseSummaryFilePath env variable
	ReleaseSummaryFilePath = "RELEASE_SUMMARY_FILE_PATH"
	// Issue Summary File path env variable
	IssueSummaryFilePath = "ISSUE_SUMMARY_FILE_PATH"
)

func readConfiguration() Configuration {

	log.Println("Reading the configuration file")
	var config Configuration
	var configFile = getEnvOrDefault(ConfigFile, "config.yaml")
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

func main() {

	config := readConfiguration()
	client := NewClient()
	log.Println("Listing repositories for each organization")

	expectedPrList, orgReleasesList, issueList, errorOccurred := getExpectedReportsLists(config, client)
	if errorOccurred {
		return
	}
	externalPRList, externalReleaseList, externalIssueList := getExternalReports(config, expectedPrList, orgReleasesList, issueList)
	var reportFilePath, templateFilePath string
	var err error

	if config.PullRequests.PRReportShouldRun {
		// Save noteworthy PRs into a file
		reportFilePath = getEnvOrDefault(PrSummaryFilePath, config.PullRequests.PRSummaryFileName)
		templateFilePath = "html/template/pr-template.html"
		err = generateReport(config.PullRequests.PRDataFile, expectedPrList, reportFilePath, templateFilePath)
		if err != nil {
			log.Fatalf("Failed to generate the report: %v, with template: %v. Error is: %v", reportFilePath, templateFilePath, err)
		}
		err = generateExternalPR(config.PullRequests.PRExternalTemplate, externalPRList)
		if err != nil {
			log.Fatalf("Failed to generate the report: %v, with template: %v. Error is: %v",
				config.PullRequests.PRExternalTemplate.Output, config.PullRequests.PRExternalTemplate.Input, err)
		}
	}

	if config.Releases.ReleaseReportShouldRun {
		// Save releases into a file
		reportFilePath = getEnvOrDefault(ReleaseSummaryFilePath, config.Releases.ReleaseSummaryFileName)
		templateFilePath = "html/template/release-template.html"
		err = generateReport(config.Releases.ReleaseDataFile, orgReleasesList, reportFilePath, templateFilePath)
		if err != nil {
			log.Fatalf("Err: %v", err)
		}
		err = generateExternalRelease(config.Releases.ReleaseExternalTemplate, externalReleaseList)
		if err != nil {
			log.Fatalf("Failed to generate the report: %v, with template: %v. Error is: %v",
				config.Releases.ReleaseExternalTemplate.Output, config.Releases.ReleaseExternalTemplate.Input, err)
		}
	}

	if config.Issues.IssueReportShouldRun {
		// Save releases into a file
		reportFilePath = getEnvOrDefault(IssueSummaryFilePath, config.Issues.IssueSummaryFileName)
		templateFilePath = "html/template/issue-template.html"
		err = generateReport(config.Issues.IssueDataFile, issueList, reportFilePath, templateFilePath)
		if err != nil {
			log.Fatalf("Err: %v", err)
		}
		err = generateExternalIssue(config.Issues.IssueExternalTemplate, externalIssueList)
		if err != nil {
			log.Fatalf("Failed to generate the report: %v, with template: %v. Error is: %v",
				config.Issues.IssueExternalTemplate.Output, config.Issues.IssueExternalTemplate.Input, err)
		}
	}
}

func getOrg(organizations []Organization, name string) OrganizationStructure {
	for _, org := range organizations {
		if org.Organization.Name == name {
			return org.Organization
		}
	}
	// Unexpected
	log.Fatalln("Unexpected organization found!")
	return OrganizationStructure{}
}

func getExternalReports(config Configuration,
	expectedPrList []PullRequestDetails,
	orgReleasesList []ReleaseDetails,
	issueList []IssueDetails,
) ([]ExternalPRDetails, []ExternalReleaseDetails, []ExternalIssueDetails) {
	if !config.GlobalConfiguration.ExternalTemplate.Enabled {
		return []ExternalPRDetails{}, []ExternalReleaseDetails{}, []ExternalIssueDetails{}
	}
	var externalPRDetails []ExternalPRDetails
	for _, org := range expectedPrList {
		organization := getOrg(config.GlobalConfiguration.Organizations, org.Organization)
		for _, repo := range org.PrRepoLists {
			elementPRDetails := ExternalPRDetails{
				Organization: OrganizationStructure{
					Github: org.Organization,
					Name:   organization.Name,
				},
				Repository: RepositoryStructure{
					Name: repo.Repository,
					Link: "https://github.com/" + org.Organization + "/" + repo.Repository,
				},
				PRs: repo.PRs,
			}
			externalPRDetails = append(externalPRDetails, elementPRDetails)
		}
	}
	var externalReleaseDetails []ExternalReleaseDetails
	for _, org := range orgReleasesList {
		organization := getOrg(config.GlobalConfiguration.Organizations, org.Organization)
		for _, repo := range org.ReleaseRepoLists {
			elementRelease := ExternalReleaseDetails{
				Organization: OrganizationStructure{
					Github: org.Organization,
					Name:   organization.Name,
				},
				Repository: RepositoryStructure{
					Name: repo.Repository,
					Link: "https://github.com/" + org.Organization + "/" + repo.Repository,
				},
				Releases: repo.Releases,
			}
			externalReleaseDetails = append(externalReleaseDetails, elementRelease)
		}
	}
	var externalIssueDetails []ExternalIssueDetails
	for _, org := range issueList {
		organization := getOrg(config.GlobalConfiguration.Organizations, org.Organization)
		for _, repo := range org.IssueLists {
			elementIssue := ExternalIssueDetails{
				Organization: OrganizationStructure{
					Github: org.Organization,
					Name:   organization.Name,
				},
				Repository: RepositoryStructure{
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

func generateExternalPR(externalTemplate ElementExternalTemplate, values []ExternalPRDetails) error {
	if len(values) == 0 {
		log.Println("External template file generation is not requested")
		return nil
	}
	for _, value := range values {
		err := generateExternalFile(value, value.Repository.Name, value.Organization.Github, externalTemplate)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateExternalIssue(externalTemplate ElementExternalTemplate, values []ExternalIssueDetails) error {
	if len(values) == 0 {
		log.Println("External template file generation is not requested")
		return nil
	}
	for _, value := range values {
		err := generateExternalFile(value, value.Repository.Name, value.Organization.Github, externalTemplate)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateExternalRelease(externalTemplate ElementExternalTemplate, values []ExternalReleaseDetails) error {
	if len(values) == 0 {
		log.Println("External template file generation is not requested")
		return nil
	}
	for _, value := range values {
		err := generateExternalFile(value, value.Repository.Name, value.Organization.Github, externalTemplate)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateExternalFile(value interface{}, filename string, org string, externalTemplate ElementExternalTemplate) error {
	var err error
	outputFileName := filename
	outputPath := path.Join(externalTemplate.Output, org)
	err = os.MkdirAll(outputPath, 755)
	if err != nil {
		return err
	}
	outputFilePath := path.Join(outputPath, outputFileName)
	err = PrettyPrint(value, outputFilePath, externalTemplate.Input)
	if err != nil {
		return err
	}
	return nil
}

func generateReport(dataFileName string, v interface{}, reportFilePath string, templateFilePath string) error {

	err := SaveIntoFile(v, dataFileName)
	if err != nil {
		log.Fatalf("Error in saving report as json : %v. Error is: %v", reportFilePath, err)
		return err
	}

	err = PrettyPrint(v, reportFilePath, templateFilePath)
	if err != nil {
		log.Fatalf("Error in generating report html: %v. Error is: %v", reportFilePath, err)
		return err
	}

	return nil
}

func getExpectedReportsLists(config Configuration, client ClientInterface) ([]PullRequestDetails, []ReleaseDetails, []IssueDetails, bool) {
	var expectedPrList []PullRequestDetails
	var orgReleasesList []ReleaseDetails
	var issueList []IssueDetails

	for _, organization := range config.GlobalConfiguration.Organizations {

		repos, err := client.ListRepositories(organization.Organization.Github)
		if err != nil {
			log.Fatalf("Err: %v", err)
			return nil, nil, nil, true
		}
		log.Printf("List for %v is : %v", organization, repos)

		//// Pull requests
		expectedPrs, errorOccurred := getExpectedPullRequests(client, organization, repos, config)
		if errorOccurred {
			return nil, nil, nil, true
		}
		expectedPrList = append(expectedPrList, expectedPrs)

		// Releases
		releaseList, errorOccurred := getReleaseList(client, organization, repos, config)
		if errorOccurred {
			return nil, nil, nil, true
		}
		orgReleasesList = append(orgReleasesList, releaseList)

		//good first issues and other configured tags
		expectedIssues, errorOccurred := getIssueList(client, organization, repos, config)
		if errorOccurred {
			return nil, nil, nil, true
		}
		issueList = append(issueList, expectedIssues)

	}
	return expectedPrList, orgReleasesList, issueList, false
}

func getReleaseList(client ClientInterface, organization Organization, repos []string, config Configuration) (ReleaseDetails, bool) {
	orgReleases, err := client.ListReleases(organization.Organization.Github, repos, config.GlobalConfiguration.DaysCount)
	if err != nil {
		log.Fatalf("Err: %v", err)
		return ReleaseDetails{}, true
	}
	releaseList := ReleaseDetails{
		Organization:     organization.Organization.Github,
		ReleaseRepoLists: orgReleases,
	}
	return releaseList, false
}

func getIssueList(client ClientInterface, organization Organization, repos []string, config Configuration) (IssueDetails, bool) {
	issues, err := client.IssueWithLabels(organization.Organization.Github, repos, config.Issues.IssueTags, config.Issues.IssueCreatedHistoryDays)
	if err != nil {
		log.Fatalf("Err: %v", err)
		return IssueDetails{}, true
	}
	issueList := IssueDetails{
		Organization: organization.Organization.Github,
		IssueLists:   issues,
	}
	return issueList, false
}

func getExpectedPullRequests(client ClientInterface, organization Organization, repos []string, config Configuration) (PullRequestDetails, bool) {
	pRs, err := client.ListPRs(organization.Organization.Github, repos, config.GlobalConfiguration.DaysCount)
	if err != nil {
		log.Fatalf("Err: %v", err)
		return PullRequestDetails{}, true
	}
	expectedPrs := PullRequestDetails{
		Organization: organization.Organization.Github,
		PrRepoLists:  pRs,
	}
	return expectedPrs, false
}

func getEnvOrDefault(env, defaultValue string) string {
	value, isPresent := os.LookupEnv(env)
	if isPresent {
		return value
	}
	return defaultValue
}
