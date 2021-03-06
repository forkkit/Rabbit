// Copyright 2019 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package model

import (
	"encoding/json"
)

// ReleaseRequest struct
type ReleaseRequest struct {
	Name    string
	URL     string
	Version string
}

// LoadFromJSON load object from json
func (r *ReleaseRequest) LoadFromJSON(data []byte) (bool, error) {
	err := json.Unmarshal(data, &r)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ConvertToJSON converts object to json
func (r *ReleaseRequest) ConvertToJSON() (string, error) {
	data, err := json.Marshal(&r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
