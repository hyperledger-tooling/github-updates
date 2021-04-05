package main

// Configuration reads the input config file
type Configuration struct {
	Organizations           []Organization `yaml:"organizations"`
	DaysCount               int            `yaml:"days"`
	IssueTags               []string       `yaml:"issue-tags"`
	FileName                string         `yaml:"filename"`
	PrSummaryFileName       string         `yaml:"pr-summary-filename"`
	ReleaseSummaryFileName  string         `yaml:"release-summary-filename"`
	IssueSummaryFileName    string         `yaml:"issue-summary-filename"`
	IssueCreatedHistoryDays int            `yaml:"issue-created-history-days"`
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
