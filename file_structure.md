```
/geo-project-go
├── cmd/                        # Główne punkty startowe aplikacji (funkcje main)
│   ├── svc-analytics/          # main.go dla T1 (pgx)
│   ├── svc-catalog/            # main.go dla T2 (goqu)
│   ├── svc-features/           # main.go dla T3 (gorm)
│   ├── svc-auth/               # main.go dla T4 (ent)
│   ├── svc-annotations/        # main.go dla T5 (mongo-driver)
│   └── svc-revisions/          # main.go dla T6 (mongo + validators)
│
├── internal/                   # Kod prywatny poszczególnych mikroserwisów
│   ├── analytics/              # Logika domenowa serwisu T1
│   ├── layers/                 # Logika domenowa serwisu T2
│   ├── features/               # Logika domenowa serwisu T3
│   ├── auth/                   # Logika domenowa serwisu T4
│   ├── annotations/                 # Logika domenowa serwisu T5
│   └── revisions/              # Logika domenowa serwisu T6
│
├── nginx/                      # Przechowuje nginx.conf             
|
│
│
├── pkg/                        # WSPÓŁDZIELONY KOD
│   ├── errors/                 # Ujednolicony format błędów
│   ├── database/               # Singletony i połączenia
│   └── middlewares/            # Wspólne middleware
│
│
├── db/                         # Skrypty bazodanowe
│   ├── migrations/             # Pliki migracji SQL(golang-migrate i Atlas)
│
├── docker-compose.yml          # Plik Compose
├── Dockerfile                  # Multi-stage Dockerfile (jeden dla wszystkich serwisów w Go)
├── .env.example
├── Makefile                    # Komendy zarządzające
```