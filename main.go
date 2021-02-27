package main

import (
	"io/ioutil"
	"log"
	"os"

	yaml "gopkg.in/yaml.v2"
)

const (
	// ConfigFile env variable that can be overridden
	ConfigFile = "CONFIG_FILE"
	// GitHubToken env variable
	GitHubToken = "GITHUB_TOKEN"
)

func readConfiguration() Configuration {

	log.Println("Reading the configuration file")
	var config Configuration
	var configFile string = getEnvOrDefault(ConfigFile, "config.yaml")
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
	for _, organization := range config.Organizations {
		repos, err := client.ListRepositories(organization.Organization.Name)
		if err != nil {
			log.Fatalf("Err: %v", err)
			return
		}
		log.Printf("List for %v is : %v", organization, repos)
		pRs, err := client.ListPRs(organization.Organization.Name, repos)
		if err != nil {
			log.Fatalf("Err: %v", err)
			return
		}
		//log.Printf("List for PRs is : %v", pRs)
		err = saveIntoFile(pRs, config.FileName)
		if err != nil {
			log.Fatalf("Err: %v", err)
			return
		}
	}
}

func getEnvOrDefault(env, defaultValue string) string {
	value, isPresent := os.LookupEnv(env)
	if isPresent {
		return value
	}
	return defaultValue
}
