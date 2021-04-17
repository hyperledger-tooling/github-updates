package main

// Configuration reads the input config file
type Configuration struct {
	GlobalConfiguration GlobalConfiguration      `yaml:"global"`
	Issues              IssueConfiguration       `yaml:"issues"`
	PullRequests        PullRequestConfiguration `yaml:"pull-requests"`
	Releases            ReleaseConfiguration     `yaml:"releases"`
}

type IssueConfiguration struct {
	IssueTags               []string `yaml:"issue-tags"`
	IssueCreatedHistoryDays int      `yaml:"created-history-days"`
	IssueSummaryFileName    string   `yaml:"summary-filename"`
	IssueReportShouldRun    bool     `yaml:"should-run"`
	IssueDataFile           string   `yaml:"data-file"`
}

type PullRequestConfiguration struct {
	PRSummaryFileName string `yaml:"summary-filename"`
	PRReportShouldRun bool   `yaml:"should-run"`
	PRDataFile        string `yaml:"data-file"`
}

type ReleaseConfiguration struct {
	ReleaseSummaryFileName string `yaml:"summary-filename"`
	ReleaseReportShouldRun bool   `yaml:"should-run"`
	ReleaseDataFile        string `yaml:"data-file"`
}

type GlobalConfiguration struct {
	Organizations []Organization `yaml:"organizations"`
	DaysCount     int            `yaml:"scrape-duration-days"`
}

// Organization represents GitHub organization
type Organization struct {
	Organization OrganizationStructure `yaml:"organization"`
}

// OrganizationStructure has information about particular
// GitHub Organization
type OrganizationStructure struct {
	Name string `yaml:"name"`
}
