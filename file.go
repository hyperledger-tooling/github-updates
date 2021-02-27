package main

import (
	"encoding/json"
	"io/ioutil"
)

func saveIntoFile(pRs []PullRequest, fileName string) error {
	fileContents, err := json.Marshal(pRs)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fileName, fileContents, 0644)
}
