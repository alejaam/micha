# Implementation Plan: Coherencia y visualización de dashboards (salir del “look onboarding”)

## Technical Context
- Repo: `/opt/data/repos/micha`
- Frontend:
  - React 18 + Vite (`frontend/package.json`)
  - Router: `react-router-dom` (router en `frontend/src/router.jsx`)
  - Estilos globales: `frontend/src/styles.css`
  - Tests: Vitest (`npm run test:coverage -- --run`)
- Onboarding actual:
  - Layout: `frontend/src/layouts/OnboardingLayout.jsx` (incluye `StepIndicator`)
  - Guard: `frontend/src/layouts/ProtectedOnboardingLayout.jsx`
  - Rutas onboarding: `/onboarding/household`, `/onboarding/cards`, `/onboarding/fixed-expenses`
- Problema UX confirmado:
  - Las pantallas accesibles desde *Settings/Rules* apuntan a rutas `/onboarding/*` (p. ej. `RulesPage.jsx` navega a `/onboarding/cards` y `/onboarding/household`).
  - Esto hace que se vean como “wizard onboarding” y en algunos casos la navegación queda “encerrada” (ej. acción principal “Saltar por ahora” en `OnboardingFixedExpensesPage.jsx`).

## Goal (from spec)
- Implementar cambios de UI/UX para que las pantallas tipo dashboard/gestión **no parezcan onboarding**.
- Asegurar navegación coherente (siempre se pueda volver) y eliminar el “dead-end” de depender de “Saltar por ahora”.

## Proposed Approach (high level)
1) **Separar “setup onboarding” vs “gestión post-onboarding”** a nivel de routing.
2) Para gestión post-onboarding, usar layout normal de la app (header + bottom nav) y copy/estructura consistentes con el resto de pantallas.
3) Para las rutas onboarding que se mantengan, asegurar:
   - siempre exista una acción explícita de “Volver” (a paso anterior o a Settings),
   - y que “Salir” no sea el único camino.

## Detailed Steps

### 1) Inventario y decisiones de routing
- Revisar rutas actuales y puntos de entrada:
  - `frontend/src/router.jsx`
  - `frontend/src/pages/RulesPage.jsx`
  - `frontend/src/layouts/ProtectedOnboardingLayout.jsx`
- Decisión de diseño:
  - **Crear rutas post-onboarding “normales”** para gestión:
    - `/cards` → gestión de tarjetas
    - `/settings/household` (o similar) → edición de ajustes del hogar
  - Mantener rutas `/onboarding/*` para primer setup, pero **dejar de usarlas como “settings”**.

### 2) Tarjetas: crear pantalla de gestión (no-onboarding)
- Crear `frontend/src/pages/CardsPage.jsx` reutilizando lógica de `OnboardingCardsPage.jsx`:
  - Listar tarjetas, seleccionar preferida, crear y eliminar.
  - Ajustar copy y estructura para modo “Settings” (sin `authEyebrow/authTitle` de onboarding).
  - Añadir navegación superior o acciones claras:
    - botón “Volver a ajustes” → `/rules`
    - opcional: “Volver al dashboard” → `/`
- Actualizar `RulesPage.jsx`:
  - Cambiar `navigate('/onboarding/cards')` → `navigate('/cards')`
- Router:
  - Añadir ruta `{ path: 'cards', element: <CardsPage /> }` bajo `AppLayout`.

### 3) Hogar: crear pantalla de ajustes (no-onboarding)
- Crear `frontend/src/pages/HouseholdSettingsPage.jsx` (o nombre equivalente) basado en `OnboardingHouseholdPage.jsx`.
  - Nota: hoy solo existe `createHousehold` en `frontend/src/api.js`; si no existe endpoint de update, el plan es:
    - (A) Si el backend soporta update y falta el cliente: agregar `updateHousehold(...)` en `api.js` y usarlo.
    - (B) Si NO hay update backend: convertir esta pantalla a “ver configuración + acciones disponibles” (o deshabilitar edición) para no prometer algo imposible.
  - Añadir navegación “Volver a ajustes” (`/rules`).
- Actualizar `RulesPage.jsx`:
  - Cambiar `navigate('/onboarding/household')` → `navigate('/settings/household')`
- Router:
  - Añadir ruta `{ path: 'settings/household', element: <HouseholdSettingsPage /> }` bajo `AppLayout`.

### 4) Gastos fijos: evitar experiencia onboarding cuando es gestión
- Ya existe `FixedExpensesPage.jsx` en `/fixed-expenses` con UI de gestión.
- Ajustes:
  - Si algún flujo post-onboarding aún navega a `/onboarding/fixed-expenses` (por ejemplo `OnboardingCardsPage.jsx` con “Continuar…”), cambiar a `/fixed-expenses` cuando el usuario ya está en modo gestión.
  - En `OnboardingFixedExpensesPage.jsx`:
    - Reemplazar el “Saltar por ahora” como única salida.
    - Añadir botón “Volver” (a `/onboarding/cards` si es wizard) y/o “Volver a ajustes” (a `/rules`).
    - Cambiar el copy de “Saltar por ahora” a algo con intención clara (ej. “Ir al dashboard” o “Volver al inicio”).

### 5) Consistencia visual: reducir clases onboarding en gestión
- Identificar y limitar uso de:
  - `.onboardingShell`, `.onboardingCard`, `.onboardingHeader`, `.dashboardOnboarding` (`frontend/src/styles.css`).
- Para pantallas de gestión (Cards, Household Settings): usar patrones existentes:
  - `pageGrid`, `card`, `listHeader`, `listTitle`, `u-text-sm`, `u-text-dim`, etc.

## Files Likely to Change
- `frontend/src/router.jsx`
- `frontend/src/pages/RulesPage.jsx`
- `frontend/src/pages/OnboardingCardsPage.jsx` (solo para ajustar navegación / dual-mode si se mantiene)
- `frontend/src/pages/OnboardingFixedExpensesPage.jsx` (arreglar navegación “no volver”)
- `frontend/src/pages/OnboardingHouseholdPage.jsx` (posible extracción/reuso de UI)
- **Nuevo:** `frontend/src/pages/CardsPage.jsx`
- **Nuevo:** `frontend/src/pages/HouseholdSettingsPage.jsx`
- `frontend/src/styles.css` (si hace falta ajustes de layout/clases)

## Test & Validation Strategy
- Frontend (mínimo):
  - `cd frontend && npm run lint`
  - `cd frontend && npm run test:coverage -- --run`
  - Smoke manual (cuando corra el frontend):
    1) Ir a Settings (`/rules`) → Gestionar tarjetas → volver a Settings
    2) Settings → Editar ajustes del hogar → volver a Settings
    3) Entrar a `/onboarding/fixed-expenses` y confirmar que existe **Volver** y/o salida clara (no solo “Saltar por ahora”)
    4) Verificar que el dashboard vacío (`DashboardPage.jsx`) no muestre UI onboarding salvo cuando realmente no hay household.
- Backend:
  - Opcional si Go está disponible en el entorno: `cd backend && go test ./...`

## Risks & Mitigations
- Riesgo: No existe endpoint para editar household (solo `createHousehold`).
  - Mitigación: confirmar backend endpoints antes de prometer “editar”; si falta, se crea pantalla de “ver + acciones” o se implementa endpoint (si entra en scope).
- Riesgo: romper flujo onboarding existente.
  - Mitigación: mantener rutas `/onboarding/*` intactas, pero cambiar solo los puntos de entrada post-onboarding (RulesPage + redirects).

## Open Questions / Needs Clarification
1) Navegación esperada: ¿prefieres que el botón principal sea “Volver a ajustes” (`/rules`) o “Volver al dashboard” (`/`), o ambos?
2) ¿Las pantallas que “se sienten onboarding” son principalmente las de Settings (Tarjetas/Hogar/Gastos fijos) o también el *empty state* del Dashboard?

---

## Gate
¿Apruebas este **plan** para que genere `tasks.md` y empiece a implementar?
- Responde **APPROVE** o **CHANGES** (con lista de cambios).