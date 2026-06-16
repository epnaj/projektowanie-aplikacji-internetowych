
### Decyzja:  

Tracking Pixel zawsze zwraca kod 200 OK, nawet przy nieaktywnym linku. Jest nie cache'owalny.
---
### Kontekst:  

Pixel jest wbudowany, dodany do cudzej (klienckiej) strony internetowej, czy dołączony do email'a. Jego jednym zadaniem jest zarejestrowanie, że ktoś wyświetlił wiadomość, czy odwiedził pewną stronę, bez żadnych ingerencji.
---
### Alternatywy  

Można zwracać 404 dla nieznanych haszów, czy 410 dla nieaktwynych linków.

---
### Uzasadnienie

Jakakolwiek odpowiedź inna niż 200 OK sprawi, że strona może się źle renderować, piksel może stać się widoczny, może też dawać po prostu informację osobie postronnej która nie powinna była wiedzieć, że jest śledzona i nie dawać jej informacji np. o statusie linku. Brak możliwości cache'owania sprawia, że każde odwiedzenie czy odświerzenie strony jest liczone.

---
### Trade-offy

Klient nie ma możliwości sprawdzenia poprawności linku Pixela, czy np. dobrze go skopiowała, nie może też się dowiedzieć czy dane są poprawnie rejestrowane w aplikacji. Szczególnie ważne jest to gdy np. programista klienta nie uprawnień do logowania się do aplikacji. Niecache'owalność ma też cenę - każde odświeżenie i każda odsłona wysyłą zapytanie do serwera aplikacji, bez odciążenia przez cache przeglądarki, więc przy dużym ruchu koszt obsługi rośnie. To cena za dokładność zliczania - pojedynczy cache'owany piksel oznaczałby niepoliczone wyświetlenia.