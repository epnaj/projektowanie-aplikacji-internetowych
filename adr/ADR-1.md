
### Decyzja:  


Go jako główny język wykorzystany w projekcie  
---
### Kontekst:  

Aplikacja musi jednocześnie udostępniać REST API, renderować prosty frontend i obsługiwać ścieżkę krytyczną pixela: bardzo wiele krótkich, równoległych zapytań z niskim czasem odpowiedzi. Język musi to łączyć z dość szybkim tempem pisania kodu i łatwym wdrożeniem w kontenerze.

---
### Alternatywy  
Rozważone było kilka języków i podejść przede wszystkim javascript, typescript - to domyślne opcje webowe, oferują sporo w kwestii API oraz frontendu, Python z Fast API - prosty do obsłużenia i pisania backend, ale bez natywnego wsparcia forontendu. Go - obsługuje domyślnie implementację REST API, prosty frontend, ma możliwość używania frameworków, sporą zaletą jest jego bardzo lekka wielowątkowość i wydajność.

---
### Uzasadnienie

Aplikacja wymaga obsłużenia potencjalnie bardzo wielu, krótkich, równoległych zapytań oferując bardzo szybki czas odpowiedzi (nie chcemy, aby prosty pixel spowalniał całą stronę).
Równocześnie dostępna jest możliwość implementacji czy późniejszej refaktoryzacji frontendu oraz natywnie wspierane budowanie API i świetna wielowątkowość. Język ma wbudowany framework testowy (go test) oraz narzędzie formatujące i sprawdzające składnię (gofmt, go vet), co redukuje liczbę zewnętrznych zależności narzędziowych

---
### Trade-offy

Rozwlekła obsługa błędów - wzorzec if err != nil powtarza się w całym kodzie i wydłuża go, jest to także zaleta, bo język został tak zaprojektowany aby zmusić programistę do obsługi tych irtyujących błędów. Ekosystem frontendowy jest dużo uboższy niż w JavaScript/TypeScript, więc bogaty UI wymaga obejść. Mniej metaprogramowania i abstrakcji niż w językach dynamicznych oznacza więcej kodu „przepisującego" dane.