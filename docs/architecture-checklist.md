# Checklist Arquitectónico

- [x] El dominio no importa paquetes de infraestructura.
- [x] Los casos de uso no dependen del framework HTTP.
- [x] Los handlers HTTP solo orquestan entrada/salida.
- [x] Los repositorios concretos implementan puertos de salida.
- [x] La inyección de dependencias ocurre en `cmd/api`.
