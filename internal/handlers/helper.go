package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type CtxSessionKey string

const CtxSessionName = CtxSessionKey("session")

type ErrorResponse struct {
	Errors string `json:"errors"`
}

type ResponseData struct {
	Session string
	Url     string
	Status  int
	Data    any
}

func WriteResponse(w http.ResponseWriter, logger *zap.SugaredLogger, responseData ResponseData) error {
	logger.Infof("session: %s; response status: %d; url: %s",
		responseData.Session,
		responseData.Status,
		responseData.Url)

	jsonData, err := json.Marshal(responseData.Data)
	if err != nil {
		return err
	}

	w.WriteHeader(responseData.Status)
	_, err = w.Write(jsonData)
	if err != nil {
		return err
	}

	return nil
}
