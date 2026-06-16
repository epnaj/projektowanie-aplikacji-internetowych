
### Decyzja:  

Brak Redisa, zamiast tego użyto własnej implementacji cache.
---
### Kontekst:  

Obsłużenie linku pixela i zwiększanie licznika dla konkrentych linków jest ścieżką krytyczną. Wymaga badzo szybkiej obsługi, ta ścieżka kodu nie może w międzyczasie czekać na aktualizację bazy danych, sama baza także nie powinna być aktualizowana przy każdym "kliknięciu" w link, może obsługiwać ich tysiące czy miliony na sekundę.
---
### Alternatywy  
Każdorazowe aktualizowanie bazy danych, użycie Redisa.

---
### Uzasadnienie

Własny bufor zbiera, a następnie okresowo zapisuje w bazie danych wszystkie zakutalizowane statystyki komunikując się z bazą danych znacznie rzadziej i nie generując zbędnych opóźnień. Redis został odrzucony ze względnu na sztuczne zwiększanie rozmiaru aplikacji - należałoby obsługiwać i utrzymywać Redisa tylko po to aby zwiększać wartość liczników. Został także rozważony jako główna baza danych, ale nie jest on przeznaczony do budowania relacyjnych baz danych, a jedyny sposób na to żeby nie tracił informacji przy np. crashu, to tworzenie regularnych zrzutów jego stanu, to po prostu za dużo pracy na tak prosty problem.Warto także mieć na uwadze, że aby obsłużyć 1000000 linków na dany okres (jedynie "kliknięte" i aktywne w trakcie danego okresu buforowego są w ogóle alokowane) potrzebne jest maksymalnie ~55MB pamięci.

Prosty pomiar zużycia pamięci w zależności od ilości linków aktywowanych w okresie:
  ┌────────┬────────────────────┬─────────┬────────────────────────────────────────────┐
  │   N    │ B/wpis             │ mln/GiB │           Gdzie w cyklu wzrostu            │
  ├────────┼────────────────────┼─────────┼────────────────────────────────────────────┤
  │ 1 mln  │ 55,81              │ 19,2    │ tuż po podwojeniu (~48% pełna, duży slack) │
  ├────────┼────────────────────┼─────────┼────────────────────────────────────────────┤
  │ 4 mln  │ 55,82              │ 19,2    │ jw.                                        │
  ├────────┼────────────────────┼─────────┼────────────────────────────────────────────┤
  │ 7 mln  │ 34,42              │ 31,2    │ tuż przed wzrostem (~95% pełna) ← minimum  │
  ├────────┼────────────────────┼─────────┼────────────────────────────────────────────┤
  │ 10 mln │ 44,73              │ 24,0    │ środek cyklu                               │
  ├────────┼────────────────────┼─────────┼────────────────────────────────────────────┤
  │ 30 mln │ 51,66              │ 20,8    │ bliżej górnej granicy                      │
  └────────┴────────────────────┴─────────┴────────────────────────────────────────────┘

---
### Trade-offy

"Kliknięcia", które już są zarejestrowane w cache, ale nie w bazie danych są stracone w przypadku crasha - to nie problem przy analityce, ale jeśli użytkownik wymaga dokładnych danych, to musi liczyć się z tym, że może ich nie otrzymać. Obecne rozwiązanie nie skaluje się dobrze wszerz, jeśli aplikacja wymagałaby skalowania czy rozporszenia koniecznym będzie zaimplementowanie cache typu Redis, na osobnych serwerach.