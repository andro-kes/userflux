# Userflux — базовая архитектура (local-only)

## Контекст
Userflux генерирует нагрузку **только через Gateway по HTTP**. Gateway внутри делает fan-out по **gRPC** в микросервисы.  
На текущем этапе всё **локально**, без Prometheus/Postgres: результаты пишутся на диск, отчёты строятся Python-скриптами.

---

## 1) Контейнерная/компонентная схема (local-only)

```mermaid
flowchart TB
  subgraph DEV["Local machine"]
    CLI["userfluxctl (CLI)"]
    D["userfluxd (Orchestrator)"]
    A["userflux-agent (Load generator)"]
    FS[("Local FS: runs/<run_id>/")]
    PY["Python analytics (plots/reports)"]
  end

  CLI -->|"start run / status"| D
  D -->|"start run (local process)"| A
  A -->|"HTTP load"| GW["Gateway (HTTP)"]
  GW -->|"gRPC fan-out"| MS["Microservices (internal gRPC)"]
  A -->|"samples + summary"| FS
  D -->|"metadata + logs"| FS
  PY -->|"read artifacts"| FS
  CLI -->|"generate report"| PY
```

---

## 2) Поток выполнения одного прогона (Run lifecycle)

```mermaid
sequenceDiagram
  autonumber
  participant C as userfluxctl
  participant D as userfluxd
  participant A as userflux-agent
  participant G as Gateway (HTTP)
  participant F as Local FS (runs/run_id)
  participant P as Python analytics

  C->>D: StartRun(config)
  D->>F: create dir + write config/status=running
  D->>A: launch agent(run_id, config)
  loop while run active
    A->>G: HTTP requests (vUsers concurrency)
    A->>F: write samples (opt) + update aggregates
  end
  A->>F: write summary + status=finished/failed
  D->>F: write orchestrator logs + final status
  C->>P: GenerateReport(run_id)
  P->>F: read summary/samples and write report
```

---

## 3) Артефакты прогона (на диске)

```mermaid
flowchart LR
  RID["run_id"] --> DIR["runs/<run_id>/"]
  DIR --> CFG["config.json"]
  DIR --> ST["status.json"]
  DIR --> SUM["summary.json"]
  DIR --> SAMP["samples.ndjson (optional)"]
  DIR --> LOG["logs.txt (optional)"]
  DIR --> REP["report.html and/or png (generated)"]
```

---

## 4) Ключевые принципы (для текущего этапа)
- Профиль нагрузки: **concurrency (virtual users)** + стадии (ramp/hold).
- Метрики на старте: **локальная агрегация** (latency p50/p95/p99, error rate, throughput).
- `summary.json` — основной контракт результата для Python-аналитики.