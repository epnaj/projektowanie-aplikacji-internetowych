
### Decyzja:  

PostgreSQL jako jedyne źródło trwałego zapisu (system of record); model relacyjny.
---
### Kontekst:  

Dane domeny są silnie powiązane relacjami właścicielskimi: użytkownik ma projekty, projekt ma linki, link ma godzinowe statystyki. Wymagana jest integralność tych relacji (kaskadowe usuwanie, unikalność e-maila i hasza linku) oraz trwałość - zliczone statystyki nie mogą znikać. Statystyki mają naturalny kształt tabelaryczny: (link, godzina, liczba odsłon). Dodatkowo aplikacja musi wstać jednym poleceniem `docker compose up`, więc baza musi dać się trywialnie uruchomić w kontenerze.

---
### Alternatywy  

SQLite - wbudowana, zero osobnego serwera, jeden plik. Słabo znosi jednak współbieżny zapis z wielu połączeń, a ścieżka pixela jest mocno współbieżna, gorzej też rozdziela aplikację od danych.
MongoDB / baza dokumentowa - wymaganie dopuszcza model dokumentowy, ale nasze dane są z natury relacyjne.  
MySQL / MariaDB - porównywalna jakościowo, ale Postgres oferuje bogatsze typy (TIMESTAMPTZ) oraz `ON CONFLICT ... DO UPDATE`, ważny przy aktualizacji statystyk.
Redis jako główna baza - odrzucony osobno (ADR-5): nie jest relacyjny, trwałość tylko przez okresowe zrzuty.

---
### Uzasadnienie

To problem dyktuje model, nie odwrotnie. Relacje właścicielskie i ograniczenia integralności (klucze obce z kaskadą, UNIQUE na e-mailu i haszu) chcemy mieć egzekwowane przez bazę, a nie ręcznie w aplikacji, bo ręczna spójność jest niewyczerpanym źródłem błędów. Postgres daje TIMESTAMPTZ na bezstrefowy znacznik godziny oraz `INSERT ... ON CONFLICT (link_id, hour) DO UPDATE SET hits = statistics.hits + EXCLUDED.hits`, który idealnie pasuje do okresowego zrzutu bufora (ADR-5) - jedno zapytanie albo tworzy kubełek, albo dolicza do istniejącego. Serwerowy charakter bazy pozwala w przyszłości oddzielić aplikację od danych bez zmiany kodu, bo adapter już siedzi za interfejsem (ADR-10). Uruchomienie w kontenerze jest trywialne - oficjalny obraz `postgres:17-alpine` z healthcheckiem.

---
### Trade-offy

Postgres to osobny proces i serwis, który trzeba uruchomić oraz utrzymać - w przeciwieństwie do SQLite nie ma opcji "zero zależności, jeden plik". Zwiększa to złożoność kompozycji (zależności startowe, healthcheck) i lokalny narzut zasobów. Przy ekstremalnym wolumenie zapisu pojedyncza instancja może stać się wąskim gardłem i będzie wymagać replikacji lub partycjonowania, czego obecne rozwiązanie nie adresuje.
