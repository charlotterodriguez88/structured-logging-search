// main.go — Exemplo de ingestão e busca de logs estruturados.
package main

import (
	"fmt"
	"time"
)

func main() {
	// Ingerir alguns logs estruturados.
	_, err := Infrai.Logs.Ingest(IngestRequest{
		Level:     "info",
		Message:   "Usuário fez login",
		Service:   "auth-service",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata:  map[string]string{"user_id": "12345", "ip": "192.168.1.1"},
	})
	if err != nil {
		fmt.Printf("Erro ao ingerir log 1: %v\n", err)
	} else {
		fmt.Println("Log 1 ingerido.")
	}

	_, err = Infrai.Logs.Ingest(IngestRequest{
		Level:     "error",
		Message:   "Falha na conexão com o banco de dados",
		Service:   "db-service",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata:  map[string]string{"db_host": "postgres.example.com", "error_code": "ECONNREFUSED"},
	})
	if err != nil {
		fmt.Printf("Erro ao ingerir log 2: %v\n", err)
	} else {
		fmt.Println("Log 2 ingerido.")
	}

	// Buscar logs por nível.
	fmt.Println("\nBuscando logs de nível 'error':")
	results, err := Infrai.Logs.Search(SearchRequest{
		Level: "error",
	})
	if err != nil {
		fmt.Printf("Erro ao buscar logs: %v\n", err)
	} else {
		for _, log := range results {
			fmt.Printf("- [%s] %s: %s (serviço: %s, metadados: %v)\n", log.Timestamp, log.Level, log.Message, log.Service, log.Metadata)
		}
	}

	// Buscar logs por serviço.
	fmt.Println("\nBuscando logs do serviço 'auth-service':")
	results, err = Infrai.Logs.Search(SearchRequest{
		Service: "auth-service",
	})
	if err != nil {
		fmt.Printf("Erro ao buscar logs: %v\n", err)
	} else {
		for _, log := range results {
			fmt.Printf("- [%s] %s: %s (metadados: %v)\n", log.Timestamp, log.Level, log.Message, log.Metadata)
		}
	}

	// Buscar logs por texto.
	fmt.Println("\nBuscando logs contendo 'falha':")
	results, err = Infrai.Logs.Search(SearchRequest{
		Query: "falha",
	})
	if err != nil {
		fmt.Printf("Erro ao buscar logs: %v\n", err)
	} else {
		for _, log := range results {
			fmt.Printf("- [%s] %s: %s (serviço: %s)\n", log.Timestamp, log.Level, log.Message, log.Service)
		}
	}

	// Exemplo com intervalo de tempo.
	fmt.Println("\nBuscando logs dos últimos 5 minutos:")
	now := time.Now().UTC()
	fiveMinAgo := now.Add(-5 * time.Minute)
	results, err = Infrai.Logs.Search(SearchRequest{
		StartTime: fiveMinAgo.Format(time.RFC3339),
		EndTime:   now.Format(time.RFC3339),
	})
	if err != nil {
		fmt.Printf("Erro ao buscar logs: %v\n", err)
	} else {
		for _, log := range results {
			fmt.Printf("- [%s] %s: %s (serviço: %s)\n", log.Timestamp, log.Level, log.Message, log.Service)
		}
	}
}
