package main

import (
	ctx "context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

// ClientInterface is for testing
type ClientInterface interface {
	ListRepositories(string) ([]string, error)
	ListPRs(string, []string, int) ([]PrList, error)
	ListReleases(string, []string, int) ([]ReleaseList, error)
	IssueWithLabels(string, []string, []string, int) ([]IssueList, error)
}

// Client is the custom handler for all requests
type Client struct {
	Client  *github.Client
	Context ctx.Context
}

// PullRequestDetails contains organization name
// and PrLists
type PullRequestDetails struct {
	Organization string   `json:"organization,omitempty"`
	PrRepoLists  []PrList `json:"prlists,omitempty"`
}

// PrList contains repository name
// and the associated PRs
type PrList struct {
	Repository string               `json:"repository,omitempty"`
	PRs        []github.PullRequest `json:"prs,omitempty"`
}

type ReleaseDetails struct {
	Organization     string        `json:"organization,omitempty"`
	ReleaseRepoLists []ReleaseList `json:"releaseList,omitempty"`
}

type IssueDetails struct {
	Organization string      `json:"organization,omitempty"`
	IssueLists   []IssueList `json:"issueLists,omitempty"`
}

type ReleaseList struct {
	Repository string                     `json:"repository,omitempty"`
	Releases   []github.RepositoryRelease `json:"releases,omitempty"`
}

type IssueList struct {
	Repository string         `json:"repository,omitempty"`
	Labels     []string       `json:"labels,omitempty"`
	Issues     []github.Issue `json:"issues,omitempty"`
}

// NewClient creates a new instance of GitHub client
func NewClient() ClientInterface {
	token := getEnvOrDefault(GitHubToken, "")
	context := ctx.Background()
	if token == "" {
		return Client{
			Client:  github.NewClient(nil),
			Context: context,
		}
	}
	oauth2Client := oauth2.NewClient(context, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))
	return Client{
		Client:  github.NewClient(oauth2Client),
		Context: context,
	}
}

// ListRepositories returns the list of all repositories
func (c Client) ListRepositories(org string) ([]string, error) {
	listOfRepositories := []string{}
	listOption := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 20,
		},
	}
	for {
		repositories, response, err :=
			c.Client.Repositories.ListByOrg(c.Context, org, listOption)
		if err != nil {
			return nil, err
		}
		log.Printf("Response: %v", response)
		if response.StatusCode != http.StatusOK {
			return nil, errors.New("Could not get the response")
		}
		for _, repository := range repositories {
			listOfRepositories = append(listOfRepositories, *repository.Name)
		}

		if response.NextPage == 0 {
			log.Println("Breaking from the loop of repositories")
			break
		}
		// assign next page
		listOption.Page = response.NextPage
	}
	return listOfRepositories, nil
}

// ListPRs returns the list of PRs for a given organization and repository
func (c Client) ListPRs(org string, repos []string, daysCount int) ([]PrList, error) {
	prListOptions := &github.PullRequestListOptions{
		State: "all",
		ListOptions: github.ListOptions{
			PerPage: 20,
		},
	}
	dayDiff := daysCount * -1
	var pullRequests []PrList

	for _, repo := range repos {
		var listPullRequests []github.PullRequest
		prDateReached := false
		for {
			prs, response, err := c.Client.PullRequests.List(c.Context, org, repo, prListOptions)
			if err != nil {
				return nil, err
			}
			log.Printf("Response: %v", response)
			if response.StatusCode != http.StatusOK {
				return nil, errors.New("Could not get the response")
			}

			for _, pr := range prs {
				timeStamp := pr.CreatedAt
				startDate := time.Now().AddDate(0, 0, dayDiff)
				log.Println("timestamp", timeStamp, "start date", startDate, " if condition", timeStamp.Before(startDate))
				if timeStamp.Before(startDate) {
					prDateReached = true
					break
				}
				if pr.ClosedAt != nil && pr.MergedAt == nil {
					continue
				}
				listPullRequests = append(listPullRequests, *pr)
			}
			if prDateReached {
				break
			}
			if response.NextPage == 0 {
				log.Println("Breaking from the loop of repositories - PRs")
				break
			}
			// assign next page
			prListOptions.Page = response.NextPage

		}
		if len(listPullRequests) != 0 {
			pullRequestElement := PrList{
				Repository: repo,
				PRs:        listPullRequests,
			}
			pullRequests = append(pullRequests, pullRequestElement)
		}

	}
	return pullRequests, nil

}

func (c Client) ListReleases(org string, repos []string, daysCount int) ([]ReleaseList, error) {

	var listReleases []ReleaseList
	dayDiff := daysCount * -1

	releaseListOptions := &github.ListOptions{PerPage: 20}

	// Iterate over the repos
	for _, repo := range repos {

		var releaseList []github.RepositoryRelease
		releaseDateReached := false

		for {

			releases, response, err := c.Client.Repositories.ListReleases(c.Context, org, repo, releaseListOptions)
			if err != nil {
				return nil, err
			}
			log.Printf("Response: %v", response)
			if response.StatusCode != http.StatusOK {
				return nil, errors.New("Could not get the response")
			}

			// For each release, break if the date is reched
			// else add the release list to the listReleases
			// and move to the next page
			for _, release := range releases {

				publishedDate := release.PublishedAt
				startDate := time.Now().AddDate(0, 0, dayDiff)
				log.Println("publishedDate", publishedDate, "start date", startDate, " if condition", publishedDate.Before(startDate))

				if publishedDate.Before(startDate) {
					releaseDateReached = true
					break
				}

				releaseList = append(releaseList, *release)
			}

			if releaseDateReached {
				break
			}
			if response.NextPage == 0 {
				log.Println("Breaking from the loop of repositories - Releases")
				break
			}
			// assign next page
			releaseListOptions.Page = response.NextPage
		}

		if len(releaseList) != 0 {
			releaseListElement := ReleaseList{
				Repository: repo,
				Releases:   releaseList,
			}
			listReleases = append(listReleases, releaseListElement)
		}
	}
	return listReleases, nil
}

func (c Client) IssueWithLabels(org string, repos []string, issueLabels []string, daysCount int) ([]IssueList, error) {
	var issueList []IssueList

	dayDiff := daysCount * -1

	//get open issues to be worked on and which has not been assigned to someone
	issueListOptions := &github.IssueListByRepoOptions{
		State:    "open",
		Assignee: "",
		Labels:   issueLabels,
		ListOptions: github.ListOptions{
			PerPage: 20,
		},
	}

	// Iterate over the repos
	for _, repo := range repos {

		var listIssues []github.Issue
		issueDateReached := false

		for {

			issues, response, err := c.Client.Issues.ListByRepo(c.Context, org, repo, issueListOptions)
			if err != nil {
				return nil, err
			}
			log.Printf("Response: %v", response)
			if response.StatusCode != http.StatusOK {
				return nil, errors.New("Could not get the response for fetching issues")
			}
			for _, issue := range issues {

				publishedDate := issue.GetCreatedAt()
				startDate := time.Now().AddDate(0, 0, dayDiff)
				log.Println("publishedDate", publishedDate, "start date", startDate, " if condition", publishedDate.Before(startDate))

				if publishedDate.Before(startDate) {
					issueDateReached = true
					break
				}
				listIssues = append(listIssues, *issue)

			}
			if issueDateReached {
				break
			}
			if response.NextPage == 0 {
				log.Println("Breaking from the loop of repositories - Issues")
				break
			}
			// assign next page
			issueListOptions.Page = response.NextPage
		}
		if len(listIssues) != 0 {
			issueElement := IssueList{
				Repository: repo,
				Labels:     issueLabels,
				Issues:     listIssues,
			}
			issueList = append(issueList, issueElement)
		}
	}
	return issueList, nil
}
