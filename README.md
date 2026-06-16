# Pixel Tracker

## Opis 
Lekki, wydajny system do śledzenia ruchu na stronach internetowych. Właściciele stron generują unikalny skrypt (pixel), który po umieszczeniu na stronie docelowej wysyła zapytania do naszego API.  
System zlicza odsłony i wizyty w czasie rzeczywistym, wykorzystując architekturę buforowania w pamięci, a dane prezentuje na interaktywnym panelu (HTMX).

[![CI](https://github.com/epnaj/projektowanie-aplikacji-internetowych/actions/workflows/ci.yml/badge.svg)](https://github.com/epnaj/projektowanie-aplikacji-internetowych/actions/workflows/ci.yml)

## Uruchamianie

Po sklonowaniu / pobraniu repozytorium wystarczy wywołać:
> docker compose up --build

Trzy usługi startują w ustalonej kolejności:

- `db` — PostgreSQL
- `migrate` — zakłada/aktualizuje schemat bazy, po czym kończy działanie
- `app` — serwer Go, startuje dopiero po udanej migracji

Aplikacja domyślnie dostępna jest pod:
> localhost:8080

## Wypełnienie danymi
Patrz:
> /seed/README.md

"Na skróty" - Z korzenia repozytorium uruchom:
> cd seed && go run . -url http://localhost:8080

Dane można sprawdzić logując się przez:
>  /login jako seed-user-01@example.test / password123

## Wymagania

| Wymaganie | Opis           | Jak / czy zostało zaimplementowane                    |
|-----------|----------------|-------------------------------------------------------|
| R1        | Backend API    | REST API z relacjami: Project, User, Link, Statistics |
| R2        | Baza danych    | PostgreSQL oraz migracje za pomocą go-migrate         |
| R3        | Frontend       | Minimalistyczny HTMX                                  |
| R4        | Autentykacja   | Bazuje jedynie na podpisanych ciasteczkach            |
| R5        | Konteneryzacja | Docker compose z healthcheck'iem bazy danych          |
| R6        | Repozytorium   | Jest publiczne                                        |

---

## Elementy dodatkowe

| Wymaganie        | Jak / czy zostało zaimplementowane                     |
|------------------|--------------------------------------------------------|
| Cache            | Własna implementacja                                   |
| Testy            | Testy jednostkowe, E2E                                 |
| CI/CD            | Dodane GitHub Actions                                  |
| Observability    | Dostępne są rotowane logi, endpointy /healthz /readyz  |
| Walidacja danych | Obecna w API                                           |
| Seed data        | Obecna w /seed; go run . -url http://localhost:8080    |

---

## Architektura

Aplikacja jest napisana w Go i podzielona na warstwy tak, żeby logika nie
zależała od konkretnej bazy danych ani od HTTP.

- **Rdzeń** (`internal/core`) — czysta logika: użytkownicy, projekty, linki,
  statystyki. Nie wie nic o bazie ani o sieci. Mówi tylko, czego potrzebuje
  (interfejsy, np. „zapisz projekt"), a nie jak to zrobić.
- **Adaptery** — konkretne implementacje tych interfejsów:
  - `internal/store/postgres` — zapis do PostgreSQL
  - `internal/store/memory` — wersja w pamięci (do testów i prostego
    uruchomienia) oraz bufor odsłon
  - `internal/auth` — hasła (bcrypt) i podpisane ciasteczka sesji
- **Warstwa web** (`web`) — serwer HTTP: REST API (JSON) oraz panel HTML z HTMX.
  Zamienia żądania na wywołania rdzenia i z powrotem.
- **Złożenie** (`cmd/server`) — jedyne miejsce, które łączy wszystko razem i
  wybiera bazę (PostgreSQL, gdy ustawiono `DATABASE_URL`; w pamięci, gdy nie).

Dzięki temu rdzeń da się testować bez bazy, a bazę można podmienić bez zmiany
logiki. Uzasadnienia wszystkich decyzji są w katalogu [`adr/`](adr/).

### Ścieżka odsłony (pixel)

To najczęstsza i najważniejsza operacja, więc jest zoptymalizowana pod szybkość:

1. Przeglądarka pobiera obrazek: `GET /pixel/{hash}`.
2. Serwer **zawsze** zwraca przezroczysty obrazek 1×1 (kod 200) — nawet gdy link
   nie istnieje lub jest wyłączony. Nie psuje to cudzej strony i nie zdradza,
   które linki są prawdziwe.
3. Odsłona trafia do bufora w pamięci, a nie od razu do bazy.
4. Co kilka sekund osobny wątek zrzuca zebrane odsłony do PostgreSQL — jednym
   zapytaniem na parę (link, godzina).

## Endpointy + API (JSON)

| Metoda i ścieżka                       | Auth | Co robi                               |
|----------------------------------------|------|---------------------------------------|
| `POST /api/register`                   | nie  | zakłada konto                         |
| `POST /api/login`                      | nie  | loguje, ustawia ciasteczko sesji      |
| `POST /api/logout`                     | tak  | wylogowuje                            |
| `GET /api/me`                          | tak  | dane zalogowanego użytkownika         |
| `GET /api/projects`                    | tak  | lista projektów                       |
| `POST /api/projects`                   | tak  | nowy projekt                          |
| `GET /api/projects/{id}`               | tak  | jeden projekt                         |
| `PATCH /api/projects/{id}`             | tak  | zmiana nazwy projektu                 |
| `DELETE /api/projects/{id}`            | tak  | usuwa projekt                         |
| `GET /api/projects/{projectId}/links`  | tak  | linki w projekcie                     |
| `POST /api/projects/{projectId}/links` | tak  | nowy link                             |
| `GET /api/links/{id}`                  | tak  | jeden link                            |
| `PATCH /api/links/{id}`                | tak  | zmiana nazwy lub włączenie/wyłączenie |
| `DELETE /api/links/{id}`               | tak  | usuwa link                            |
| `GET /api/projects/{projectId}/stats`  | tak  | statystyki całego projektu            |
| `GET /api/links/{id}/stats`            | tak  | statystyki jednego linku              |
| `GET /pixel/{hash}`                    | nie  | zlicza odsłonę, zwraca obrazek 1×1    |
| `GET /healthz`                         | nie  | czy serwis żyje                       |
| `GET /readyz`                          | nie  | czy baza danych odpowiada             |

### Panel HTML (HTMX)

| Ścieżka                          | Co robi                                    |
|----------------------------------|--------------------------------------------|
| `GET /`                          | przekierowanie do panelu                   |
| `GET /login`, `GET /register`    | formularze logowania i rejestracji         |
| `GET /app`                       | panel z listą projektów                    |
| `GET /app/projects/{id}`         | szczegóły projektu i jego linki            |
| pozostałe `POST/DELETE /app/...` | akcje panelu (twórz/usuń projekt lub link) |
