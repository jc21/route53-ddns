package model

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
)

// AWSConfig is the settings that are saved for use in updating
type AWSConfig struct {
	AWSKeyID          string `survey:"aws_key_id"`
	AWSKeySecret      string `survey:"aws_key_secret"`
	ZoneID            string `survey:"zone_id"`
	Recordset         string `survey:"recordset"`
	PushoverUserToken string `survey:"pushover_user_token"`
}

func (c *AWSConfig) Write(filename string) error {
	content, _ := json.MarshalIndent(c, "", " ")

	// Make sure the ".aws" folder exists
	folder := path.Dir(filename)
	dirErr := os.MkdirAll(folder, os.ModePerm)
	if dirErr != nil {
		return dirErr
	}

	err := ioutil.WriteFile(filename, content, 0600)
	if err != nil {
		return err
	}

	return nil
}
