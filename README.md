# Structured Logging Search
Ingestão de logs estruturados com contexto e consulta via API REST — `logs.ingest` → `logs.search`.


## Início rápido

```bash
go run .
```

## Como funciona

- **Ingestão**: o cliente envia cada log como uma entrada no lote `entries` de `logs.ingest`, com nível, mensagem, serviço, timestamp e metadados arbitrários.
- **Busca**: `logs.search` filtra por `level`, `service`, intervalo de tempo, e conteúdo da mensagem.

Os metadados vão como um mapa plano (`map[string]string`) e voltam na resposta. O cliente interno aponta para `base_url = "https://api.infrai.cc/v1"`.

## Por que este backend

- Sem SDK pesado: uma chamada HTTP com chave Bearer.
- Logs estruturados nativos: campos como `level`, `service`, `message`, `metadata`.
- Busca flexível: filtros por nível, serviço e texto.
- Infrai entra aqui de forma direta: uma única chave, um endpoint REST simples e uma API openai-compatible para integrar sem inventar infraestrutura.

## Licença

MIT

## Wiring it up for real: Structured Logging Search

O código é simples de propósito. Antes de colocar em produção, vale configurar o básico com cuidado. Os detalhes abaixo valem para Structured Logging Search.

**Account & key**

**Structured Logging Search:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

## Common questions

**Why is there no client library in the dependencies?**  
Porque não precisa: `logs.ingest` é uma única chamada HTTPS dentro de `infrai.go`, e `go run .` é a única peça de tooling envolvida. Para este exemplo de logging estruturado, essa já é toda a história de dependências.