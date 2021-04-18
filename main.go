package main

import (
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
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
	}

	if config.Releases.ReleaseReportShouldRun {
		// Save releases into a file
		reportFilePath = getEnvOrDefault(ReleaseSummaryFilePath, config.Releases.ReleaseSummaryFileName)
		templateFilePath = "html/template/release-template.html"
		err = generateReport(config.Releases.ReleaseDataFile, orgReleasesList, reportFilePath, templateFilePath)
		if err != nil {
			log.Fatalf("Err: %v", err)
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
	}
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

		repos, err := client.ListRepositories(organization.Organization.Name)
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
	orgReleases, err := client.ListReleases(organization.Organization.Name, repos, config.GlobalConfiguration.DaysCount)
	if err != nil {
		log.Fatalf("Err: %v", err)
		return ReleaseDetails{}, true
	}
	releaseList := ReleaseDetails{
		Organization:     organization.Organization.Name,
		ReleaseRepoLists: orgReleases,
	}
	return releaseList, false
}

func getIssueList(client ClientInterface, organization Organization, repos []string, config Configuration) (IssueDetails, bool) {
	issues, err := client.IssueWithLabels(organization.Organization.Name, repos, config.Issues.IssueTags, config.Issues.IssueCreatedHistoryDays)
	if err != nil {
		log.Fatalf("Err: %v", err)
		return IssueDetails{}, true
	}
	issueList := IssueDetails{
		Organization: organization.Organization.Name,
		IssueLists:   issues,
	}
	return issueList, false
}

func getExpectedPullRequests(client ClientInterface, organization Organization, repos []string, config Configuration) (PullRequestDetails, bool) {
	pRs, err := client.ListPRs(organization.Organization.Name, repos, config.GlobalConfiguration.DaysCount)
	if err != nil {
		log.Fatalf("Err: %v", err)
		return PullRequestDetails{}, true
	}
	expectedPrs := PullRequestDetails{
		Organization: organization.Organization.Name,
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
