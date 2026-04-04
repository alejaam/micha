# Micha

Aplicación de finanzas para gastos compartidos (roomies, parejas y familias).

## Stack inicial
- Backend: Go + DDD + Clean Architecture + Hexagonal
- Frontend: React + Vite
- Base de datos: PostgreSQL 16
- Deploy: Docker Compose en VPS propio

## Estructura
- `backend/`: API y lógica de negocio
- `frontend/`: SPA React
- `Dockerfile`: build principal del backend
- `docker-compose.yml`: orquestación local
- `docs/`: principios y ADRs

## Estado actual
Primera iteración de scaffolding arquitectónico.
