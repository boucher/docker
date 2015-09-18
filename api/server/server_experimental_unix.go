// +build experimental,!windows

package server

import (
	"encoding/json"
	"fmt"
	"github.com/docker/docker/pkg/version"
	"github.com/docker/docker/runconfig"
	"net/http"
)

func addExperimentalRoutes(s *Server, m map[string]map[string]HTTPAPIFunc) {
	m["POST"]["/containers/{name:.*}/checkpoint"] = s.postContainersCheckpoint
	m["POST"]["/containers/{name:.*}/restore"] = s.postContainersRestore
}

func (s *Server) registerSubRouter() {
	httpHandler := s.daemon.NetworkAPIRouter()

	subrouter := s.router.PathPrefix("/v{version:[0-9.]+}/networks").Subrouter()
	subrouter.Methods("GET", "POST", "PUT", "DELETE").HandlerFunc(httpHandler)
	subrouter = s.router.PathPrefix("/networks").Subrouter()
	subrouter.Methods("GET", "POST", "PUT", "DELETE").HandlerFunc(httpHandler)

	subrouter = s.router.PathPrefix("/v{version:[0-9.]+}/services").Subrouter()
	subrouter.Methods("GET", "POST", "PUT", "DELETE").HandlerFunc(httpHandler)
	subrouter = s.router.PathPrefix("/services").Subrouter()
	subrouter.Methods("GET", "POST", "PUT", "DELETE").HandlerFunc(httpHandler)
}

func (s *Server) postContainersCheckpoint(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	if err := parseForm(r); err != nil {
		return err
	}

	criuOpts := &runconfig.CriuConfig{}
	if err := json.NewDecoder(r.Body).Decode(criuOpts); err != nil {
		return err
	}

	if err := s.daemon.ContainerCheckpoint(vars["name"], criuOpts); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) postContainersRestore(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	if err := parseForm(r); err != nil {
		return err
	}

	restoreOpts := runconfig.RestoreConfig{}
	if err := json.NewDecoder(r.Body).Decode(&restoreOpts); err != nil {
		return err
	}

	if err := s.daemon.ContainerRestore(vars["name"], &restoreOpts.CriuOpts, restoreOpts.ForceRestore); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
