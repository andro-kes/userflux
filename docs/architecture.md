# Userflux — базовая архитектура (local-only)

## Контекст
Userflux генерирует нагрузку **только через Gateway по HTTP**. Gateway внутри делает fan-out по **gRPC** в микросервисы.  
На текущем этапе всё **локально**, без Prometheus/Postgres: результаты пишутся на диск, отчёты строятся Python-скриптами.

---

## 1) Контейнерная/компонентная схема (local-only)

```mermaid
flowchart TB
  subgraph DEV["Local machine"]
    H["userflux (CLI)"]
    O["Orchestrator"]
    A["Agent (Load generator)"]
    W["Writer (Results collector)"]
    FS[("File System")]
    subgraph ARTIFACTS["Artifacts"]
      SESS["sessions/"]
      RES["results/"]
      LOGS["logs/"]
    end
  end

  H -->|"parse args"| O
  O -->|"read script"| SCRIPTS["scripts/*.yml"]
  O -->|"create session"| A
  A -->|"spawn vUsers"| GW["Gateway (HTTP)"]
  GW -->|"gRPC fan-out"| MS["Microservices"]
  A -->|"success/fail signals"| W
  W -->|"write JSON"| RES
  O -->|"copy config"| SESS
  H -->|"write logs"| LOGS
```

---

## 2) Поток выполнения одного прогона (Run lifecycle)

```mermaid
sequenceDiagram
  autonumber
  participant CLI as userflux (main)
  participant O as Orchestrator
  participant A as Agent
  participant W as Writer
  participant GW as Gateway (HTTP)
  participant FS as File System

  CLI->>O: Orchestrator(script, logger)
  O->>FS: Read scripts/<script>.yml
  O->>FS: Create sessions/session_N
  O->>FS: Create results/result_N
  O->>A: RunAgent(session)
  A->>W: Start writer goroutine
  
  loop while duration not exceeded
    A->>A: RandomDelay (exp/uniform/normal)
    A->>A: Spawn user goroutine
    A->>GW: Execute HTTP flow
    alt success
      A->>W: success channel
    else failure
      A->>W: fail channel
    end
  end
  
  A->>A: Wait for all goroutines
  A->>W: Context cancelled
  W->>FS: Encode final JSON to result_N
  CLI->>CLI: Exit
```

---

## 3) Структура проекта

```
userflux/
├── cmd/
│   └── main.go                 # CLI entry point
├── internal/
│   ├── agent/
│   │   ├── agent.go            # Load generation logic
│   │   ├── helpers.go          # Random delay, body generation
│   │   └── writer.go           # Result aggregation
│   ├── logging/
│   │   └── logger.go           # Multi-writer logger (file + stderr)
│   ├── orchestrator/
│   │   └── orchestrator.go     # Script parsing, session management
│   └── session/
│       └── models.go           # Data structures
├── scripts/                    # YAML test scenarios
├── sessions/                   # Session configs (auto-created)
├── results/                    # JSON results (auto-created)
├── logs/                       # Log files (auto-created)
└── docs/
    └── architecture.md         # This file
```

---

## 4) Артефакты прогона (на диске)

```mermaid
flowchart LR
  RUN["Run N"] --> SESS["sessions/session_N"]
  RUN --> RES["results/result_N"]
  RUN --> LOG["logs/userflux.log"]
  
  SESS --> |"contains"| YAML["Script YAML copy"]
  RES --> |"contains"| JSON["Success/Failure counts"]
  LOG --> |"contains"| LOGS["Timestamped execution logs"]
```

### Формат result_N:
```json
{
  "Script": "register",
  "Total": 150,
  "Success": 148,
  "Failure": 2
}
```

---

## 5) Стратегии задержки между запросами

Agent поддерживает три режима генерации задержки (`RandomDelay`):

| Режим | Описание | Параметры |
|-------|----------|-----------|
| `uniform` | Равномерное распределение ±X% от base | `jitterFraction` = доля (0.3 = ±30%) |
| `exp` | Экспоненциальное распределение | `base` = среднее время |
| `normal` | Нормальное распределение (Box-Muller) | `jitterFraction` = стандартное отклонение как доля от base |

Все режимы поддерживают `min` и `max` для ограничения значений.

---

## 6) Ключевые принципы

- **Профиль нагрузки**: concurrency через goroutines + рандомизированные интервалы
- **Метрики**: локальная агрегация (Total, Success, Failure) через атомарные счётчики
- **Каналы**: успех/неудача передаются через buffered channels в Writer
- **Graceful shutdown**: контекст с таймаутом + ожидание всех goroutines через WaitGroup
- **Логирование**: dual-output (файл + stderr) с timestamps и source location

---

## 7) Текущие ограничения

- Нет поддержки ramp-up/ramp-down стадий
- Параметр `users` в конфиге пока не используется (запланировано)
- Нет расширенных метрик (latency p50/p95/p99, throughput)
- Python-аналитика ещё не реализована
- Результаты не включают временные метки отдельных запросов