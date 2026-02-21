# Checklist Arquitectónico

- [ ] El dominio no importa paquetes de infraestructura.
- [ ] Los casos de uso no dependen del framework HTTP.
- [ ] Los handlers HTTP solo orquestan entrada/salida.
- [ ] Los repositorios concretos implementan puertos de salida.
- [ ] La inyección de dependencias ocurre en `cmd/api`.
