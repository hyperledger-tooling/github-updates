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
	"github.com/google/go-github/v33/github"
)

type RepositoryStructure struct {
	Name string
	Link string
}

type ExternalPRDetails struct {
	Organization OrganizationStructure
	Repository   RepositoryStructure
	PRs          []github.PullRequest
}

type ExternalIssueDetails struct {
	Organization OrganizationStructure
	Repository   RepositoryStructure
	Issues       []github.Issue
}

type ExternalReleaseDetails struct {
	Organization OrganizationStructure
	Repository   RepositoryStructure
	Releases     []github.RepositoryRelease
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
