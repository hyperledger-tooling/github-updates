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

package utils

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

func GetEnvOrDefault(env, defaultValue string) string {
	value, isPresent := os.LookupEnv(env)
	if isPresent {
		return value
	}
	return defaultValue
}
