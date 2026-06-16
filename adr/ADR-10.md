
### Decyzja:  

Architektura heksagonalna (porty i adaptery). Logika domenowa (pakiet core) jest czysta i nie importuje żadnej infrastruktury; zależności są zdefiniowane jako interfejsy (porty) w `core`, a konkretne implementacje (adaptery) wstrzykiwane w jedynym miejscu kompozycji `cmd/server/main.go`.
---
### Kontekst:  

Aplikacja ma kilka wymiennych elementów infrastruktury: magazyn w pamięci (do testów i prostego uruchomienia), baza danych PostgreSQL (produkcyjny), bufor w pamięci oraz hasher bcrypt. Chcemy testować reguły domenowe szybko i bez bazy, a także móc podmienić bazę bez przepisywania logiki. Bez wyraźnej granicy logika i SQL splatają się i każdy test wymaga prawdziwej bazy.

---
### Alternatywy  

Logika rozmawiająca bezpośrednio z konkretną bazą - trzeba pisać mniej kodu na starcie, ale reguły domenowe nie dają się testować bez bazy, a zmiana bazy dotyka całego kodu.
Nieco bardziej przemawiającą alternatywą może być pełna czysta architektura z podziałem na koncentryczne warstwy: encje, use-case'y, adaptery interfejsu i frameworki. Daje najmocniejszą izolację, ale nakłada konkretny narzut pracy, który przy tej skali nie zwraca się wartością:
- Każda operacja staje się osobnym use-case'em (struktura wejścia + struktura wyjścia + wykonawca), zamiast zwykłej metody serwisu. Cztery zasoby (użytkownik, projekt, link, statystyka) razy operacje CRUD to kilkadziesiąt takich obiektów.
- Reguła zależności wymusza osobny model danych na każdej granicy: encja domeny != wiersz bazy != DTO wejścia use-case'a != DTO wyjścia != DTO HTTP. To cztery-pięć niemal identycznych struktur na zasób i tyleż funkcji mapujących między nimi.
- Presenter tłumaczyłby wynik use-case'a na model widoku, ale tu "widokiem" jest po prostu fragment HTMX albo JSON DTO, więc presenter dublowałby istniejące pakiety `view` i `dto`.
Nasze reguły biznesowe są cienkie (CRUD + sprawdzenie własności + bufor), więc te warstwy w większości tylko przepuszczałyby dane, dodając kod bez realnej wartości. Pełne use-case'y i presentery zaczęłyby się opłacać dopiero, gdyby reguły urosły (złożone polityki, wiele kanałów wejścia/wyjścia); przy obecnym zakresie heksagon daje dobrą izolację, bez mnożenia warstw.

---
### Uzasadnienie

Wymienne magazyny danych oraz szybkie, hermetyczne testy - prowadzi to do portów i adapterów. `core` definiuje interfejsy `UserRepository`, `ProjectRepository`, `LinkRepository`, `StatisticRepository`, `HitBuffer`, `PasswordHasher`, a usługi operują wyłącznie na tych interfejsach. Dzięki temu testy jednostkowe domeny można testować bez instancji bazy danych, a ten sam kod produkcyjnie dostaje adapter Postgres - przełączany jedną zmienną `DATABASE_URL`, bez dotykania logiki. Granica wymusza też, że błędy infrastruktury (np. kod 23505 z Postgresa) są tłumaczone na błędy domeny (`ErrConflict`) wewnątrz adaptera, więc `core` nie wie nic o SQL-u. Wybrano heksagon, a nie pełną Clean Architecture, bo daje 90% korzyści (testowalność, wymienność) przy minimum warstw - adekwatnie do skali.

---
### Trade-offy

Więcej kodu i pośrednictwa niż przy bezpośrednim SQL-u w handlerze - dla każdej encji trzeba zdefiniować interfejs, zmock'owany i prawdziwy adapter. Pojawia się ryzyko "przeciekania abstrakcji", gdy typ specyficzny dla bazy przedostaje się do interfejsu portu.