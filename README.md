# projektowanie-aplikacji-internetowych

## Opis 
Lekki, wydajny system do śledzenia ruchu na stronach internetowych. Właściciele stron generują unikalny skrypt (pixel), który po umieszczeniu na stronie docelowej wysyła zapytania do naszego API.  
System zlicza odsłony i wizyty w czasie rzeczywistym, wykorzystując architekturę buforowania w pamięci (Redis), a dane prezentuje na interaktywnym panelu (HTMX).