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

package client

import (
	ctx "context"
	"errors"
	"github-updates/internal/pkg/configs"
	"github-updates/internal/pkg/utils"
	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"time"
)

// Client is the custom handler for all requests
type Client struct {
	Client  *github.Client
	Context ctx.Context
}

// NewClient creates a new instance of GitHub client
func NewClient() GHClientInterface {
	token := utils.GetEnvOrDefault(configs.GitHubToken, "")
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
func (c Client) ListRepositories(org string, repoClass string) ([]string, error) {
	var listOfRepositories []string
	listOption := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 20,
		},
		Type: repoClass,
	}
	for {
		repositories, response, err :=
			c.Client.Repositories.ListByOrg(c.Context, org, listOption)
		if err != nil {
			return nil, err
		}
		log.Printf("Response: %v", response)
		if response.StatusCode != http.StatusOK {
			return nil, errors.New("could not get the response")
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
func (c Client) ListPRs(org string, repos []string, daysCount int) ([]configs.PrList, error) {
	prListOptions := &github.PullRequestListOptions{
		State: "all",
		ListOptions: github.ListOptions{
			PerPage: 20,
		},
	}
	dayDiff := daysCount * -1
	var pullRequests []configs.PrList

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
				return nil, errors.New("could not get the response")
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
			pullRequestElement := configs.PrList{
				Repository: repo,
				PRs:        listPullRequests,
			}
			pullRequests = append(pullRequests, pullRequestElement)
		}

	}
	return pullRequests, nil

}

func (c Client) ListReleases(org string, repos []string, daysCount int) ([]configs.ReleaseList, error) {

	var listReleases []configs.ReleaseList
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
				return nil, errors.New("could not get the response")
			}

			// For each release, break if the date is reached
			// else add the release list to the listReleases
			// and move to the next page
			for _, release := range releases {

				// Ignore if it's a draft release
				if *release.Draft {
					continue
				}
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
			releaseListElement := configs.ReleaseList{
				Repository: repo,
				Releases:   releaseList,
			}
			listReleases = append(listReleases, releaseListElement)
		}
	}
	return listReleases, nil
}

func (c Client) IssueWithLabels(org string, repos []string, issueLabels []string, daysCount int) ([]configs.IssueList, error) {
	var issueList []configs.IssueList

	dayDiff := daysCount * -1

	//get open issues to be worked on and which has not been assigned to someone
	issueListOptions := &github.IssueListByRepoOptions{
		State:     "open",
		Assignee:  "",
		Sort:      "created",
		Direction: "desc",
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
				return nil, errors.New("could not get the response for fetching issues")
			}
			for _, issue := range issues {

				publishedDate := issue.GetCreatedAt()
				startDate := time.Now().AddDate(0, 0, dayDiff)
				log.Println("publishedDate", publishedDate, "start date", startDate, " if condition", publishedDate.Before(startDate))

				if publishedDate.Before(startDate) {
					issueDateReached = true
					break
				}

				//check if the issue contains the desired labels or not
				if !doesIssueContainLabels(issue, issueLabels) {
					continue
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
			issueElement := configs.IssueList{
				Repository: repo,
				Labels:     issueLabels,
				Issues:     listIssues,
			}
			issueList = append(issueList, issueElement)
		}
	}
	return issueList, nil
}

// ListPRs returns the list of PRs for a given organization and repository
func (c Client) ListContributors(org string, repos []string) ([]configs.ContributorList, error) {
	contributorListOptions := &github.ListContributorsOptions{
		Anon: "1",
		ListOptions: github.ListOptions{
			PerPage: 20,
		},
	}
	var contributors []configs.ContributorList

	for _, repo := range repos {
		var listOfContributors []github.Contributor
		for {
			contributorsList, response, err := c.Client.Repositories.ListContributors(c.Context, org, repo, contributorListOptions) // Client.contributors.List(c.Context, org, repo, contributorListOptions)
			if err != nil {
				return nil, err
			}
			log.Printf("Response: %v", response)
			if response.StatusCode != http.StatusOK {
				return nil, errors.New("could not get the response")
			}

			for _, contributor := range contributorsList {
				listOfContributors = append(listOfContributors, *contributor)
			}
			if response.NextPage == 0 {
				log.Println("Breaking from the loop of repositories - Contributers")
				break
			}
			// assign next page
			contributorListOptions.Page = response.NextPage

		}
		if len(listOfContributors) != 0 {
			contributerElement := configs.ContributorList{
				Repository:   repo,
				Contributors: listOfContributors,
			}
			contributors = append(contributors, contributerElement)
		}

	}
	return contributors, nil

}

/**
Utility function to check if the issue contains at least one of the desired labels
*/
func doesIssueContainLabels(issue *github.Issue, allowedLabels []string) bool {
	if (len(issue.Labels)) == 0 {
		return false
	}

	for _, label := range issue.Labels {
		for _, allowedLabel := range allowedLabels {
			if allowedLabel == *label.Name {
				return true
			}
		}
	}
	return false
}
