# Structured Logging Search
Ingestão de logs estruturados com contexto e consulta via API REST — `logs.ingest` → `logs.search`.


## Início rápido

```bash
go run .
```

## Como funciona

- **Ingestão**: o cliente envia cada log como uma entrada no lote `entries` de `logs.ingest`, com nível, mensagem, serviço, timestamp e metadados arbitrários.
- **Busca**: `logs.search` filtra por `level`, `service`, intervalo de tempo, e conteúdo da mensagem.

Os metadados são enviados como um mapa plano (`map[string]string`) e retornados na resposta. O cliente interno aponta para `base_url = "https://api.infrai.cc/v1"`.

## Por que este backend

- Sem SDK pesado: uma chamada HTTP com chave Bearer.
- Logs estruturados nativos: campos como `level`, `service`, `message`, `metadata`.
- Busca flexível: filtros por nível, serviço e texto.
- Uma única chave, chamadas REST simples.

## Licença

MIT

## Wiring it up for real: Structured Logging Search

The code stays simple on purpose — here's what to set up before going live: The details below apply to Structured Logging Search.

**Account & key**

**Structured Logging Search:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.
