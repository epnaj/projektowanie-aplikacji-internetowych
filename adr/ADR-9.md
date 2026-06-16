
### Decyzja:  

Schemat bazy zarządzany wersjonowanymi plikami migracji (`NNNNNN_nazwa.up.sql` / `.down.sql`), stosowanymi przez narzędzie golang-migrate uruchamiane jako osobny serwis w docker-compose, przed startem aplikacji.
---
### Kontekst:  

Schemat bazy będzie ewoluował, a każde wdrożenie musi otrzymać przewidywalny, powtarzalny stan bazy. Aplikacja wstaje przez docker compose i nie może wystartować, zanim schemat nie jest gotowy. Potrzebna jest ścieżka zarówno w górę (nowa wersja), jak i w dół (wycofanie), oraz pełna historia zmian w repozytorium.

---
### Alternatywy  

Tworzenie schematu w kodzie aplikacji przy starcie (`CREATE TABLE IF NOT EXISTS ...`) jest proste, ale niewersjonowane, bez ścieżki wstecz, miesza odpowiedzialności (aplikacja zarządza DDL) i przy wielu instancjach prowadzi do race condition.
Biblioteka migracji wkompilowana w aplikację (golang-migrate jako import, goose, atlas) - daje wersjonowanie, ale łamie zasadę czystości zależności (ADR-2), bo wciąga obce narzędzie do kodu Go, dodaje zależność migracji od tego, czy aplikacja działa.
Ręczne stosowanie SQL-a przez operatora - niepowtarzalne i podatne na błąd ludzki.

---
### Uzasadnienie

golang-migrate uruchomione jako osobny obraz kontenera daje wersjonowanie migracji w plikach `.sql` trzymanych w repo, ścieżkę w górę i w dół oraz brak importów w kodzie Go.  Migracje to oddzielne narzędzie, a nie część aplikacji. W docker-compose usługa "migrate" zależy od tego, czy baza danych jest "zdrowa", a aplikacja zależy od pomyślnego zakończenia migracji, więc kolejność jest wymuszona deklaratywnie. Migracja po zastosowaniu kończy działanie i nie zajmuje dodatkowych zasobów.

---
### Trade-offy

Wprowadza trzeci serwis do kompozycji i dodatkowy obraz do pobrania. Migracje pisane surowym SQL-em nie są sprawdzane przez kompilator Go - literówka w DDL ujawni się dopiero przy uruchomieniu, nie przy budowaniu. Przy migracji przerwanej w połowie golang-migrate oznacza bazę jako "dirty" i wymaga ręcznej interwencji, zamiast automatycznego rollbacku i wymaga interwencji programisty.
