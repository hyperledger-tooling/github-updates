package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"text/template"
)

func SaveIntoFile(v interface{}, fileName string) error {
	fileContents, err := json.MarshalIndent(v, "", "")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fileName, fileContents, 0644)
}

func PrettyPrint(v interface{}, fileName string, templateFile string) error {

	t, err := template.ParseFiles(templateFile)
	if err != nil {
		return err
	}

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	err = t.Execute(f, v)
	return err
}
