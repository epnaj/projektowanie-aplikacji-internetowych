
### Decyzja:  

Bezstanowe ciasteczka oparte na HMAC; brak sesji po stronie serwera
---
### Kontekst:  

Jako jedno z wymagań aplikacja miała mieć funkcjonalność logowania / potwierdzenia tożsamości.
---
### Alternatywy  
Jako alternatywa mogłaby być zastosowana typowa sesja, czy JWT.

---
### Uzasadnienie

Tak podpisane ciasteczko jest niepodrabialne i łatwo je obsłużyć zachowując w mocy poprzednie decyzje architektoniczne (brak dodatkowych frameworków / bibliotek). Dodatkowo nie wymaga to dodatkowego użycia bazy danych do przechowywania stanu użytkowników, szczególnie, że tego typu aplikacja nie potrzebuje specjalnego śledzenia stanu użytkownika na stronie (np. przy  platformie zakupowej jest to wręcz konieczne).

---
### Trade-offy

Ciasteczko jest jedynym źródłem prawdy - serwer nie weryfikuje dodatkowych informacji, jeśli ciastko jest skradzione i wciąż jest aktywne, jest to słaby punkt jeśli chodzi o bezpieczeństwo danych użytkowników.