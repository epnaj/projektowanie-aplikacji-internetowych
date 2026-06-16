
### Decyzja:  

Łączenie błędów na granicy HTTP. Brak uprawnień do cudzego zasobu (`ErrNotOwner`) zwraca 404, a nie 403; nieudane logowanie (zły login lub złe hasło) zwraca jednolite 401 bez rozróżnienia przyczyny.
---
### Kontekst:  

Identyfikatory zasobów (projekty, linki) są sekwencyjne, a e-maile użytkowników bywają znane lub zgadywalne. Standardowe kody odpowiedzi wyciekają informację: 403 na cudzy projekt mówi "ten zasób istnieje, ale nie jest twój", a osobne komunikaty "nie ma takiego użytkownika" oraz "złe hasło" pozwalają wyliczać, które konta istnieją.

---
### Alternatywy  

Zwracać dokładne kody: 403 dla cudzego zasobu, 404 dla nieistniejącego, osobne komunikaty przy logowaniu - bardziej poprawne znaczeniowo i wygodniejsze  przy debugowaniu, ale nieszczelne.
Rozróżniać przyczyny wewnętrznie, lecz ukrywać je za ogólnym komunikatem przy zachowaniu różnych kodów, informacje wciąż wyciekają przez sam kod statusu.

---
### Uzasadnienie

Zagrożeniem jest ustalanie, które identyfikatory zasobów i które konta istnieją, przez obserwację różnic w odpowiedziach. Spłaszczając "nie twoje" do "nie istnieje" (404), uniemożliwiamy atakującemu odróżnienie cudzego projektu od nieistniejącego, więc zgadywanie ID przestaje być możliwe. Jednolite 401 przy logowaniu sprawia, że "nie ma konta" i "złe hasło" są nie do odróżnienia, więc nie da się zbierać listy istniejących e-maili. Ważne, że to złączenie żyje w jednym miejscu - funkcji mapującej błąd domeny na kod HTTP na granicy web - podczas gdy domena nadal operuje precyzyjnymi błędami przydatnymi w debugowaniu (`ErrNotOwner`, `ErrNotFound`, `ErrInvalidCredentials`). Bezpieczeństwo nie zaśmieca logiki, a logi po stronie serwera wciąż pokazują co dokładnie się dzieje.

---
### Trade-offy

Trudniejszy debugging dla uczciwego klienta i programisty: "404" na zasób, który faktycznie istnieje (tylko należy do kogoś innego), bywa mylące, a jednolite 401 utrudnia zorientowanie się, czy pomylono e-mail, czy hasło. Semantyka HTTP jest świadomie nagięta (404 zamiast 403), co może zaskoczyć programistę czytającego samą specyfikację. Ochrona dotyczy enumeracji przez treść i kod odpowiedzi; nie zamyka drogi na inne ataki, jak różnice w czasie odpowiedzi, czy kradzież ciastka, które wymagałyby osobnego obsłużenia.
