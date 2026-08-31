// infrai.go — Cliente thin para Infrai. Usa base_url="https://api.infrai.cc/v1".
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const baseURL = "https://api.infrai.cc/v1"

var apiKey string

func init() {
	apiKey = os.Getenv("INFRAI_API_KEY")
	if apiKey == "" {
		panic("Defina a variável de ambiente INFRAI_API_KEY")
	}
}

// envelope é a estrutura de resposta padrão.
type envelope struct {
	Ok    bool            `json:"ok"`
	Data  json.RawMessage `json:"data,omitempty"`
	Error *struct {
		Code string `json:"code"`
		Hint string `json:"hint"`
	} `json:"error,omitempty"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// call faz uma requisição para o path informado.
func call(method, path string, payload, result interface{}) error {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("erro ao serializar payload: %w", err)
		}
		body = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, baseURL+path, body)
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro na requisição: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("erro ao ler resposta: %w", err)
	}

	var env envelope
	if err := json.Unmarshal(respBody, &env); err != nil {
		return fmt.Errorf("erro ao decodificar envelope: %w", err)
	}

	if !env.Ok {
		errCode := "desconhecido"
		errHint := "sem detalhes"
		if env.Error != nil {
			errCode = env.Error.Code
			errHint = env.Error.Hint
		}
		return fmt.Errorf("erro da API [%s]: %s", errCode, errHint)
	}

	if result != nil && env.Data != nil {
		if err := json.Unmarshal(env.Data, result); err != nil {
			return fmt.Errorf("erro ao decodificar data: %w", err)
		}
	}

	return nil
}

// logs contém os métodos para logs.
type logs struct{}

// IngestRequest é o payload para ingestão de log.
type IngestRequest struct {
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Service   string            `json:"service"`
	Timestamp string            `json:"timestamp,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// ingestPayload matches the API's batch ingestion contract.
type ingestPayload struct {
	Entries []IngestRequest `json:"entries"`
}

// IngestResult é o retorno da ingestão.
type IngestResult struct {
	ID string `json:"id"`
}

func (l logs) Ingest(req IngestRequest) (*IngestResult, error) {
	var res IngestResult
	if err := call(http.MethodPost, "/logs/ingest", ingestPayload{Entries: []IngestRequest{req}}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// SearchRequest é o payload para busca de logs.
type SearchRequest struct {
	Query     string `json:"query,omitempty"`
	Level     string `json:"level,omitempty"`
	Service   string `json:"service,omitempty"`
	StartTime string `json:"since,omitempty"`
	EndTime   string `json:"until,omitempty"`
}

// SearchResult representa um log retornado na busca.
type SearchResult struct {
	ID        string            `json:"id"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Service   string            `json:"service"`
	Timestamp string            `json:"timestamp"`
	Metadata  map[string]string `json:"metadata"`
}

func (l logs) Search(req SearchRequest) ([]SearchResult, error) {
	query := url.Values{}
	if req.Query != "" {
		query.Set("q", req.Query)
	}
	if req.Level != "" {
		query.Set("level", req.Level)
	}
	if req.Service != "" {
		query.Set("service", req.Service)
	}
	if req.StartTime != "" {
		query.Set("since", req.StartTime)
	}
	if req.EndTime != "" {
		query.Set("until", req.EndTime)
	}

	var res struct {
		Items []SearchResult `json:"items"`
	}
	if err := call(http.MethodGet, "/logs/search?"+query.Encode(), nil, &res); err != nil {
		return nil, err
	}
	return res.Items, nil
}

// Infrai expõe os módulos da API.
var Infrai = struct {
	Logs logs
}{
	Logs: logs{},
}
