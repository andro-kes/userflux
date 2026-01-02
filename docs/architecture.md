# Userflux — базовая архитектура (local-only)

## Контекст
Userflux генерирует нагрузку **только через Gateway по HTTP**. Gateway внутри делает fan-out по **gRPC** в микросервисы.  
На текущем этапе всё **локально**, без Prometheus/Postgres: результаты пишутся на диск, отчёты строятся Python-скриптами.

---

## 1) Контейнерная/компонентная схема (local-only)

```mermaid
flowchart TB
  subgraph DEV["Local machine"]
    H["userflux"]
    O["userfluxo (Orchestrator)"]
    A["userflux-agent (Load generator)"]
    FS[("CSV")]
    PY["Python analytics (plots/reports)"]
  end

  H -->|"start run"| O
  O -->|"choose script"| A
  A -->|"HTTP load"| GW["Gateway (HTTP)"]
  GW -->|"gRPC fan-out"| MS["Microservices (internal gRPC)"]
  A -->|"samples + summary"| FS
  O -->|"metadata + logs"| FS
  PY -->|"read artifacts"| FS
  H -->|"generate report"| PY
```

---

## 2) Поток выполнения одного прогона (Run lifecycle)

```mermaid
sequenceDiagram
  autonumber
  participant H as userflux
  participant O as userfluxo
  participant A as userflux-agent
  participant GW as Gateway (HTTP)
  participant FS as Local FS (runs/run_id)
  participant PY as Python analytics

  H->>O: StartRun(config)
  O->>FS: Status running...
  O->>A: launch agent(run_id, config)
  loop while run active
    A->>GW: HTTP requests (vUsers concurrency)
    A->>FS: write samples (opt) + update aggregates
  end
  A->>FS: write summary + status=finished/failed
  O->>FS: write orchestrator logs + final status
  H->>PY: GenerateReport(run_id)
  PY->>FS: read summary/samples and write report
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