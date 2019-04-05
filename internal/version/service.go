// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package version

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"go.uber.org/zap"

	"storj.io/storj/internal/sync2"
)

// Config contains the necessary Information to check the Software Version
type Config struct {
	ServerAddress  string        `help:"server address to check its version against" default:"https://version.alpha.storj.io"`
	RequestTimeout time.Duration `help:"Request timeout for version checks" default:"0h1m0s"`
	CheckInterval  time.Duration `help:"Interval to check the version" default:"0h15m0s"`
}

// Service contains the information and variables to ensure the Software is up to date
type Service struct {
	config  Config
	info    Info
	service string

	Loop *sync2.Cycle

	checked sync2.Fence
	mu      sync.Mutex
	allowed bool
}

// NewService creates a Version Check Client with default configuration
func NewService(config Config, info Info, service string) (client *Service) {
	return &Service{
		config:  config,
		info:    info,
		service: service,
		Loop:    sync2.NewCycle(config.CheckInterval),
		allowed: true,
	}
}

// CheckVersion checks to make sure the version is still okay, returning an error if not
func (srv *Service) CheckVersion(ctx context.Context) error {
	if !srv.checkVersion(ctx) {
		return fmt.Errorf("outdated software version (%v), please update", srv.info.Version.String())
	}
	return nil
}

// CheckProcessVersion is not meant to be used for peers but is meant to be
// used for other utilities
func CheckProcessVersion(ctx context.Context, config Config, info Info, service string) error {
	return NewService(config, info, service).CheckVersion(ctx)
}

// Run logs the current version information
func (srv *Service) Run(ctx context.Context) error {
	if !srv.checked.Released() {
		err := srv.CheckVersion(ctx)
		if err != nil {
			return err
		}
	}
	return srv.Loop.Run(ctx, func(ctx context.Context) error {
		srv.checkVersion(ctx)
		return nil
	})
}

// IsAllowed returns whether if the Service is allowed to operate or not
func (srv *Service) IsAllowed() bool {
	srv.checked.Wait()
	srv.mu.Lock()
	defer srv.mu.Unlock()
	return srv.allowed
}

// CheckVersion checks if the client is running latest/allowed code
func (srv *Service) checkVersion(ctx context.Context) (allowed bool) {
	defer mon.Task()(&ctx)(nil)

	defer func() {
		srv.mu.Lock()
		srv.allowed = allowed
		srv.mu.Unlock()
		srv.checked.Release()
	}()

	if !srv.info.Release {
		return true
	}

	accepted, err := srv.queryVersionFromControlServer(ctx)
	if err != nil {
		// Log about the error, but dont crash the service and allow further operation
		zap.S().Errorf("Failed to do periodic version check: ", err)
		return true
	}

	list := getFieldString(&accepted, srv.service)
	zap.S().Debugf("allowed versions from Control Server: %v", list)

	if list == nil {
		zap.S().Errorf("Empty List from Versioning Server")
		return true
	}
	if containsVersion(list, srv.info.Version) {
		zap.S().Infof("running on version %s", srv.info.Version.String())
		return true
	}
	zap.S().Errorf("running on not allowed/outdated version %s", srv.info.Version.String())
	return false
}

// QueryVersionFromControlServer handles the HTTP request to gather the allowed and latest version information
func (srv *Service) queryVersionFromControlServer(ctx context.Context) (ver AllowedVersions, err error) {
	defer mon.Task()(&ctx)(&err)

	// Tune Client to have a custom Timeout (reduces hanging software)
	client := http.Client{
		Timeout: srv.config.RequestTimeout,
	}

	// New Request that used the passed in context
	req, err := http.NewRequest("GET", srv.config.ServerAddress, nil)
	if err != nil {
		return AllowedVersions{}, err
	}
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return AllowedVersions{}, err
	}

	defer func() { _ = resp.Body.Close() }()

	err = json.NewDecoder(resp.Body).Decode(&ver)
	return ver, err
}

// DebugHandler returns a json representation of the current version information for the binary
func DebugHandler(w http.ResponseWriter, r *http.Request) {
	j, err := Build.Marshal()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(append(j, '\n'))
	if err != nil {
		zap.S().Errorf("error writing data to client %v", err)
	}
}

func getFieldString(array *AllowedVersions, field string) []SemVer {
	r := reflect.ValueOf(array)
	f := reflect.Indirect(r).FieldByName(field).Interface()
	result, ok := f.([]SemVer)
	if ok {
		return result
	}
	return nil
}
