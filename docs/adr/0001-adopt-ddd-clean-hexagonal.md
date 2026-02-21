# ADR-0001: Adoptar DDD + Clean + Hexagonal

## Estado
Aceptado

## Contexto
Necesitamos una base mantenible para una app financiera transaccional con despliegue propio en VPS.

## Decisión
- Usar DDD para modelar reglas del negocio.
- Usar Clean Architecture para controlar dependencias.
- Usar Hexagonal para aislar infraestructura por puertos/adaptadores.

## Consecuencias
- Aumenta disciplina inicial de diseño.
- Mejora testabilidad y reemplazo de adaptadores.
- Facilita evolución del dominio sin acoplarse a frameworks.
