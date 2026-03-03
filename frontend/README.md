# Frontend

Frontend mínimo en React + Vite para probar el backend de gastos.

## Requisitos

- Node.js 20+
- Backend corriendo en `http://localhost:8080`

## Ejecutar en desarrollo

```bash
cd frontend
npm install
npm run dev
```

La app levanta en `http://localhost:5173` y usa proxy de Vite para llamar al backend:

- `GET /health`
- `POST /v1/expenses`
- `GET /v1/expenses?household_id=...`
- `PATCH /v1/expenses/{id}`
- `DELETE /v1/expenses/{id}`

## Build

```bash
cd frontend
npm run build
npm run preview
```
