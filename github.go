package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

// ClientInterface is for testing
type ClientInterface interface {
	ListRepositories(string) ([]string, error)
}

// Client is the custom handler for all requests
type Client struct {
	Client *github.Client
}

// NewClient creates a new instance of GitHub client
func NewClient() ClientInterface {
	token := getEnvOrDefault(GitHubToken, "")
	if token == "" {
		return Client{
			Client: github.NewClient(nil),
		}
	}
	context := context.Background()
	oauth2Client := oauth2.NewClient(context, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))
	return Client{
		Client: github.NewClient(oauth2Client),
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
	context := context.Background()
	for {
		repositories, response, err :=
			c.Client.Repositories.ListByOrg(context, org, listOption)
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
