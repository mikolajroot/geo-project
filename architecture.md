```
/geo-project-go
├── cmd/                        # Główne punkty startowe aplikacji (funkcje main)
│   ├── api-gateway/            # main.go dla Gatewaya
│   ├── svc-analytics/          # main.go dla T1 (pgx)
│   ├── svc-catalog/            # main.go dla T2 (goqu)
│   ├── svc-features/           # main.go dla T3 (gorm)
│   ├── svc-auth/               # main.go dla T4 (ent)
│   ├── svc-ingest/             # main.go dla T5 (mongo-driver)
│   └── svc-revisions/          # main.go dla T6 (mongo + validators)
│
├── internal/                   # Kod prywatny poszczególnych mikroserwisów
│   ├── gateway/                # Logika routingu i kompensacji (Saga)
│   ├── analytics/              # Logika domenowa serwisu T1
│   ├── catalog/                # Logika domenowa serwisu T2
│   ├── features/               # Logika domenowa serwisu T3
│   ├── auth/                   # Logika domenowa serwisu T4
│   ├── ingest/                 # Logika domenowa serwisu T5
│   └── revisions/              # Logika domenowa serwisu T6
│
├── pkg/                        # WSPÓŁDZIELONY KOD
│   ├── errors/                 # Ujednolicony format błędów {error, code, details}
│   ├── database/               # Singletony i połączenia
│   └── middlewares/            # Wspólne middleware
│
├── api/                        # Kontrakty i specyfikacje
│   └── openapi.yaml            # Publikowalna specyfikacja
│
├── db/                         # Skrypty bazodanowe
│   ├── migrations/             # Pliki migracji SQL(golang-migrate i Atlas)
│   └── seeds/                  # Skrypty do seedowania (seeder_dev.go, seeder_test.go)
│
├── docker-compose.yml          # Główny plik spinający (Bazy danych + wszystkie 7 serwisów)
├── Dockerfile                  # Multi-stage Dockerfile (jeden dla wszystkich serwisów w Go)
├── .env.example                # Wymaganie
├── Makefile                    # Komendy zarządzające (make up, make seed, make migrate)
├── README.md                   # (architektura, instrukcje, diagram przepływu)
├── go.mod                      # Plik zależności Go
└── go.sum                      # Plik z hashami zależności
```