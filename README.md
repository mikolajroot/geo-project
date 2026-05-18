# Geo-project

Geo-project to system informacji geograficznej (GIS) oparty na architekturze mikroserwisów, zaprojektowany do zarządzania warstwami wektorowymi, obiektami przestrzennymi oraz adnotacjami użytkowników. System wykorzystuje architekturę hybrydową baz danych, łączącą relacyjną bazę PostgreSQL (z rozszerzeniem PostGIS) oraz dokumentową bazę MongoDB.

## Uruchamianie

1. Skopiuj plik szablonu zmiennych środowiskowych:
   `cp .env.example .env`
2. Uzupełnij zmienne środowiskowe w pliku `.env`.
3. Upewnij się, że w katalogu `secrets/` znajdują się odpowiednie pliki z poświadczeniami (`db_user.txt`, `db_password.txt`, `secret_key.txt`).
4. Uruchom system w terminalu przy pomocy zdefiniowanego skryptu:
   `make up`
   (Polecenie to wykonuje w tle `docker compose up --build -d`, uruchamiając kontenery wraz z odpowiednimi mechanizmami healthcheck).

## Architektura i Mikroserwisy

System składa się z niezależnych usług, komunikujących się w wewnętrznej sieci Docker, ukrytych za pojedynczym punktem wejścia (Nginx).

- `svc-layers` - przechowuje warstwy geograficzne. Komunikuje się z bazą PostgreSQL przy użyciu narzędzia Goqu.
- `svc-features` - przechowuje obiekty przestrzenne (Features) i ich geometrie, powiązane z warstwami. Zbudowany z wykorzystaniem ORM GORM.
- `svc-annotations` - przechowuje notatki i adnotacje terenowe powiązane z obiektami przestrzennymi. Wykorzystuje MongoDB.
- `svc-revisions` - serwis historii zmian obiektów. Implementuje symulację zachowań biblioteki Mongoose w Go (silna walidacja struktur, osadzone subdokumenty, metody pre-hook).
- `svc-auth` - odpowiada za autoryzację, zarządzanie użytkownikami oraz sesjami (tokeny JWT). Oparty na frameworku Ent.
- `svc-analytics` - dedykowany mikroserwis do odpytywania i analizy danych przestrzennych.
- `nginx` - pełni rolę API Gateway, kierując ruch zewnętrzny do odpowiednich serwisów w oparciu o ścieżki URL.

## Zagrożenia i Bezpieczeństwo

- Walidacja wejścia: Każdy serwis wykorzystuje rygorystyczne walidatory wbudowane w struktury języka Go (pakiet validator/v10). Niepoprawne geometrie, braki identyfikatorów lub nieautoryzowane zapytania są odrzucane ze statusem HTTP 400/401/403.
- Brak wycieków Stack Trace: Serwisy implementują ustandaryzowany system błędów zdefiniowany w pakiecie `pkg/errors`. Każdy błąd wewnętrzny (np. Panic lub błąd SQL) jest przechwytywany i mapowany na jednolity obiekt JSON bez ujawniania szczegółów implementacyjnych klientowi.
- Integralność Danych: Próba usunięcia obiektu (Feature) w momencie istnienia powiązanych adnotacji zostaje zablokowana kodem HTTP 409 Conflict.
- Bezpieczeństwo zapytań: Użycie nowoczesnych sterowników gwarantuje korzystanie z zapytań parametryzowanych, eliminując podatności na SQL Injection oraz NoSQL Injection.