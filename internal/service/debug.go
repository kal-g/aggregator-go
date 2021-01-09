package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/rs/zerolog"
)

func (s *Service) DebugSetLogLevel(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	bodyJSON := map[string]interface{}{}
	err = json.Unmarshal(body, &bodyJSON)
	if err != nil {
		panic(err)
	}
	logLevel := 0
	// Get log level
	if n, logLevelSet := bodyJSON["logLevel"]; logLevelSet {
		nNum, isNum := n.(float64)
		if !isNum {
			// TODO Return error
			return
		}
		logLevel = int(nNum)
		if logLevel < int(zerolog.TraceLevel) || logLevel > int(zerolog.Disabled) {
			// TODO Return error
			return
		}
		typedLogLevel := zerolog.Level(logLevel)
		zerolog.SetGlobalLevel(typedLogLevel)
		logger.Info().Msgf("Set log level to %s", typedLogLevel.String())
		w.Header().Set("Content-Type", "application/json")
	}
}
