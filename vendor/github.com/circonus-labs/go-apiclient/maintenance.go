// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Maintenance window API support - Fetch, Create, Update, Delete, and Search
// See: https://login.circonus.com/resources/api/calls/maintenance

package apiclient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/circonus-labs/go-apiclient/config"
	"github.com/pkg/errors"
)

// Maintenance defines a maintenance window. See https://login.circonus.com/resources/api/calls/maintenance for more information.
type Maintenance struct {
	Severities interface{} `json:"severities,omitempty"` // []string NOTE can be set with CSV string or []string
	CID        string      `json:"_cid,omitempty"`       // string
	Item       string      `json:"item,omitempty"`       // string
	Notes      string      `json:"notes,omitempty"`      // string
	Type       string      `json:"type,omitempty"`       // string
	Tags       []string    `json:"tags,omitempty"`       // [] len >= 0
	Start      uint        `json:"start,omitempty"`      // uint
	Stop       uint        `json:"stop,omitempty"`       // uint
}

// NewMaintenanceWindow returns a new Maintenance window (with defaults, if applicable)
func NewMaintenanceWindow() *Maintenance {
	return &Maintenance{}
}

// FetchMaintenanceWindow retrieves maintenance [window] with passed cid.
func (a *API) FetchMaintenanceWindow(cid CIDType) (*Maintenance, error) {
	if cid == nil || *cid == "" {
		return nil, errors.New("invalid maintenance window CID (none)")
	}

	var maintenanceCID string
	if !strings.HasPrefix(*cid, config.MaintenancePrefix) {
		maintenanceCID = fmt.Sprintf("%s/%s", config.MaintenancePrefix, *cid)
	} else {
		maintenanceCID = *cid
	}

	matched, err := regexp.MatchString(config.MaintenanceCIDRegex, maintenanceCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid maintenance window CID (%s)", maintenanceCID)
	}

	result, err := a.Get(maintenanceCID)
	if err != nil {
		return nil, errors.Wrap(err, "fetching maitenance window")
	}

	if a.Debug {
		a.Log.Printf("fetch maintenance window, received JSON: %s", string(result))
	}

	window := &Maintenance{}
	if err := json.Unmarshal(result, window); err != nil {
		return nil, errors.Wrap(err, "parsing maintenance window")
	}

	return window, nil
}

// FetchMaintenanceWindows retrieves all maintenance [windows] available to API Token.
func (a *API) FetchMaintenanceWindows() (*[]Maintenance, error) {
	result, err := a.Get(config.MaintenancePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "fetching maintenance windows")
	}

	var windows []Maintenance
	if err := json.Unmarshal(result, &windows); err != nil {
		return nil, errors.Wrap(err, "parsing maintenance windows")
	}

	return &windows, nil
}

// UpdateMaintenanceWindow updates passed maintenance [window].
func (a *API) UpdateMaintenanceWindow(cfg *Maintenance) (*Maintenance, error) {
	if cfg == nil {
		return nil, errors.New("invalid maintenance window config (nil)")
	}

	maintenanceCID := cfg.CID

	matched, err := regexp.MatchString(config.MaintenanceCIDRegex, maintenanceCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid maintenance window CID (%s)", maintenanceCID)
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if a.Debug {
		a.Log.Printf("update maintenance window, sending JSON: %s", string(jsonCfg))
	}

	result, err := a.Put(maintenanceCID, jsonCfg)
	if err != nil {
		return nil, errors.Wrap(err, "parsing maintenance window")
	}

	window := &Maintenance{}
	if err := json.Unmarshal(result, window); err != nil {
		return nil, err
	}

	return window, nil
}

// CreateMaintenanceWindow creates a new maintenance [window].
func (a *API) CreateMaintenanceWindow(cfg *Maintenance) (*Maintenance, error) {
	if cfg == nil {
		return nil, errors.New("invalid maintenance window config (nil)")
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if a.Debug {
		a.Log.Printf("create maintenance window, sending JSON: %s", string(jsonCfg))
	}

	result, err := a.Post(config.MaintenancePrefix, jsonCfg)
	if err != nil {
		return nil, errors.Wrap(err, "creating maintenance window")
	}

	window := &Maintenance{}
	if err := json.Unmarshal(result, window); err != nil {
		return nil, errors.Wrap(err, "parsing maintenance window")
	}

	return window, nil
}

// DeleteMaintenanceWindow deletes passed maintenance [window].
func (a *API) DeleteMaintenanceWindow(cfg *Maintenance) (bool, error) {
	if cfg == nil {
		return false, errors.New("invalid maintenance window config (nil)")
	}
	return a.DeleteMaintenanceWindowByCID(CIDType(&cfg.CID))
}

// DeleteMaintenanceWindowByCID deletes maintenance [window] with passed cid.
func (a *API) DeleteMaintenanceWindowByCID(cid CIDType) (bool, error) {
	if cid == nil || *cid == "" {
		return false, errors.New("invalid maintenance window CID (none)")
	}

	var maintenanceCID string
	if !strings.HasPrefix(*cid, config.MaintenancePrefix) {
		maintenanceCID = fmt.Sprintf("%s/%s", config.MaintenancePrefix, *cid)
	} else {
		maintenanceCID = *cid
	}

	matched, err := regexp.MatchString(config.MaintenanceCIDRegex, maintenanceCID)
	if err != nil {
		return false, err
	}
	if !matched {
		return false, errors.Errorf("invalid maintenance window CID (%s)", maintenanceCID)
	}

	_, err = a.Delete(maintenanceCID)
	if err != nil {
		return false, errors.Wrap(err, "deleting maintenance window")
	}

	return true, nil
}

// SearchMaintenanceWindows returns maintenance [windows] matching
// the specified search query and/or filter. If nil is passed for
// both parameters all maintenance [windows] will be returned.
func (a *API) SearchMaintenanceWindows(searchCriteria *SearchQueryType, filterCriteria *SearchFilterType) (*[]Maintenance, error) {
	q := url.Values{}

	if searchCriteria != nil && *searchCriteria != "" {
		q.Set("search", string(*searchCriteria))
	}

	if filterCriteria != nil && len(*filterCriteria) > 0 {
		for filter, criteria := range *filterCriteria {
			for _, val := range criteria {
				q.Add(filter, val)
			}
		}
	}

	if q.Encode() == "" {
		return a.FetchMaintenanceWindows()
	}

	reqURL := url.URL{
		Path:     config.MaintenancePrefix,
		RawQuery: q.Encode(),
	}

	result, err := a.Get(reqURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "searching maintenance windows")
	}

	var windows []Maintenance
	if err := json.Unmarshal(result, &windows); err != nil {
		return nil, errors.Wrap(err, "parsing maintenance windows")
	}

	return &windows, nil
}
