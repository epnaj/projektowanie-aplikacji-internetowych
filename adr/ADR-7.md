
### Decyzja:  

Frontend jest renderowany po stronie serwera jako HTML z fragmentami HTMX, obsługiwany przez ten sam proces i kod co REST API.
---
### Kontekst:  

Interfejs aplikacji to panel CRUD za logowaniem: listy i formularze projektów, linków oraz tabele statystyk. Nie ma wymagań bogatej interaktywności po stronie klienta.

---
### Alternatywy  
Np. Single Page Application framework, czy osobny framework działający po stronie serwera typu React / Next.js

---
### Uzasadnienie

HTMX zwraca HTML do użytkownika, wszystko dzieje się po stronie serwera, nie ma stanów klienta, ciężkiego javascript'u, ani kłopotu z wersjonowaniem API. Zachowanie jest identyczne jak wywołanie API.

---
### Trade-offy

Mało atrakcyjny wizualnie UI. 
Każda interakcja to żądanie do serwera, więc więcej rund sieciowych niż w Single Page Architecture.