package model

import (
	"encoding/json"
	"os"
	"path"
	"time"
)

// Route53State is the settings that are saved in the state file
type Route53State struct {
	ZoneID         string    `survey:"zone_id"`
	Recordset      string    `survey:"recordset"`
	LastIPv4       string    `survey:"last_ipv4"`
	LastIPv6       string    `survey:"last_ipv6"`
	LastUpdateTime time.Time `survey:"last_update_time"`
}

func (c *Route53State) Write(filename string) error {
	content, _ := json.MarshalIndent(c, "", " ")

	// Make sure the ".aws" folder exists
	folder := path.Dir(filename)
	// nolint: gosec
	dirErr := os.MkdirAll(folder, os.ModePerm)
	if dirErr != nil {
		return dirErr
	}

	err := os.WriteFile(filename, content, 0600)
	if err != nil {
		return err
	}

	return nil
}
