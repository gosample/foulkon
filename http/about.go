package http

import (
	"net/http"

	"github.com/Tecsisa/foulkon/api"
	"github.com/julienschmidt/httprouter"
)

// RESPONSE

type LoggerConfig struct {
	Type          string `json:"type,omitempty"`
	Level         string `json:"level,omitempty"`
	FileDirectory string `json:"directory,omitempty"`
}

type DatabaseConfig struct {
	Type         string `json:"type,omitempty"`
	IdleConns    int    `json:"idleconns,omitempty"`
	MaxOpenConns int    `json:"maxopenconns,omitempty"`
	ConnTtl      int    `json:"connttl,omitempty"`
}

type AuthConnectorConfig struct {
	Type          string             `json:"type,omitempty"`
	OidcProviders []api.OidcProvider `json:"oidcProviders,omitempty"`
}

type Config struct {
	Logger        LoggerConfig        `json:"logger,omitempty"`
	Database      DatabaseConfig      `json:"database,omitempty"`
	AuthConnector AuthConnectorConfig `json:"authenticator,omitempty"`
	Version       string              `json:"version,omitempty"`
}

// HANDLER

func (wh *WorkerHandler) HandleGetCurrentConfig(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Process request
	requestInfo, _, err := wh.processHttpRequest(r, w, ps, nil)
	if err != nil {
		wh.processHttpResponse(r, w, requestInfo, nil, err, http.StatusBadRequest)
		return
	}

	// Only admin is authorized
	if !requestInfo.Admin {
		err = &api.Error{
			Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
			Message: "Unauthorized, user is not admin",
		}
		wh.processHttpResponse(r, w, requestInfo, nil, err, http.StatusForbidden)
		return
	}

	wc := wh.worker.Config
	// Get Logger config
	logger := LoggerConfig{
		Type:          wc.LoggerType,
		Level:         wc.LoggerLevel,
		FileDirectory: wc.FileDirectory,
	}

	// Get Database config
	db := DatabaseConfig{
		Type:         wc.DBType,
		IdleConns:    wc.IdleConns,
		MaxOpenConns: wc.MaxOpenConns,
		ConnTtl:      wc.ConnTtl,
	}

	// Get Authenticator config
	auth := AuthConnectorConfig{
		Type:          wc.AuthType,
		OidcProviders: wc.OidcProviders,
	}

	// Config Response
	response := Config{
		Logger:        logger,
		Database:      db,
		AuthConnector: auth,
		Version:       wc.Version,
	}

	wh.processHttpResponse(r, w, requestInfo, response, nil, http.StatusOK)
}
