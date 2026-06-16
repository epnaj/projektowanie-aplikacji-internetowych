
### Decyzja:  

Całość kodu Go musi być napisana używając standardowej biblioteki, modułów golang.org/x/ oraz modułu pgx.
---
### Kontekst:  

Nieznane, pomniejsze bliblioteki i frameworki oferują sporą wygodę, ale także wprowadzają niepotrzebne zależności i kompilują projekt. Jeśli nie jest to potrzebne oraz nie sprawia, że "wymyślamy koło na nowo", nie należy niepotrzebnie korzystać z zewnętrznych bibliotek, dodaje to pracy przy sprawdzaniu wymagań FOSS i sprawia, że wdrażanie nowych członków zespołu jest dłuższe i bardziej kosztowne.

---
### Alternatywy  

Rozważono backendowe frameworki i biblioteki: Gin (routing i middleware HTTP), gorilla/sessions (obsługa sesji), go-playground/validator (walidacja pól w strukturach).

---
### Uzasadnienie

Go udostępnia wiele narzędzi natywnie, bez konieczności zwracania się ku frameworkom, dodatkowo tworzenie zależności dla oszczędności kilku linijek kodu wydaje się być strzelaniem z armaty do wróbla. Pakiety golang.org/x/ są rozszerzeniem języka wspieranym bezpośrednio przez jego twórców, będą łatane i ulepszane dopóki sam Go będzie wspierany. Moduł pgx został wykorzystany, ponieważ mimo, że nie jest niezbędny do obsługi bazy danych (tu PostgreSQL), oferuje on większą wydajność i dostęp do QueryPool.

---
### Trade-offy

Brzydszy interface - frameworki oferują znacznie ładniejszy frontend.   
Więcej kodu własnego zamiast gotowego - ręczny routing na net/http.ServeMux, własna walidacja i samodzielna obsługa sesji zamiast użycia gotowego rozwiązania. Brak dostępu do gotowego, sprawdzonego
middleware z ekosystemu. Dłuższa pierwsza implementacja i ryzyko własnych błędów w rzeczach, które dojrzała biblioteka ma już przetestowane (np. walidacja, podpisywanie sesji).