# Principios de Arquitectura

## DDD
- El dominio vive en `internal/domain`.
- Entidades y reglas de negocio no dependen de frameworks.
- Los casos de uso representan acciones del negocio.

## Clean Architecture
- Regla de dependencia: de afuera hacia adentro nunca.
- `adapters` depende de `ports` y `application`.
- `application` depende de `domain` y `ports`.
- `domain` no depende de ninguna capa interna.

## Hexagonal
- Puertos de entrada: contratos de casos de uso (`ports/inbound`).
- Puertos de salida: contratos para infraestructura (`ports/outbound`).
- Adaptadores primarios: HTTP, CLI, eventos.
- Adaptadores secundarios: PostgreSQL, mensajería, cache.
