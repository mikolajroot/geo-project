```
internal/features/
├── handlers/                   # Kontrolery HTTP
│   └── feature_handler.go      
├── services/                   # Logika biznesowa (np. walidacja geometrii przed zapisem)
│   └── feature_service.go      
├── repositories/               # Warstwa danych 
│   └── feature_repo.go         
└── models/                     # Struktury danych
    └── feature.go
```