# Tasks: Coherencia y visualización de dashboards (salir del “look onboarding”)

> Basado en:
> - `spec.md`
> - `plan.md`

## Trazabilidad (resumen)
- Objetivo principal: separar pantallas de **gestión** (post-onboarding) de pantallas **wizard** (onboarding).
- Navegación esperada: el CTA de regreso por defecto es **“Volver a ajustes” → `/rules`**.

---

## Checklist

- [ ] **T001 — Auditoría de rutas y puntos de entrada**
  - [ ] Confirmar navegación actual desde `RulesPage.jsx` hacia `/onboarding/*`.
  - [ ] Identificar cualquier otro lugar que navegue a rutas onboarding para “settings”.
  - Output: lista de lugares (archivo:línea) en comentario del PR/commit.

- [ ] **T002 — Crear pantalla de gestión de tarjetas (`/cards`)**
  - [ ] Crear `frontend/src/pages/CardsPage.jsx` reutilizando la lógica de `OnboardingCardsPage.jsx`.
  - [ ] UI tipo “gestión”: usar `listHeader/listTitle` y copy sin “Primeros pasos”.
  - [ ] Añadir acciones claras:
    - [ ] Botón primario o link: **Volver a ajustes** → `/rules`.
  - [ ] Asegurar estados: loading/empty/error coherentes.

- [ ] **T003 — Enrutar CardsPage en AppLayout**
  - [ ] Añadir ruta `/cards` en `frontend/src/router.jsx` bajo `AppLayout`.
  - [ ] Verificar que NO use `OnboardingLayout`.

- [ ] **T004 — Corregir Settings/Rules para no mandar a onboarding**
  - [ ] En `frontend/src/pages/RulesPage.jsx`, cambiar:
    - [ ] tarjetas: `/onboarding/cards` → `/cards`
    - [ ] hogar: `/onboarding/household` → `/settings/household`
  - [ ] (Opcional) cambiar texto/labels para reflejar “Ajustes” en vez de “Onboarding”.

- [ ] **T005 — Crear pantalla de ajustes del hogar (`/settings/household`) sin look onboarding**
  - [ ] Crear `frontend/src/pages/HouseholdSettingsPage.jsx`.
  - [ ] Mostrar configuración actual del household (al menos: nombre, moneda, closing day, period frequency, settlement mode), con diseño consistente.
  - [ ] Navegación: **Volver a ajustes** → `/rules`.
  - [ ] Si NO hay API de update:
    - [ ] mostrar la info + CTA “Configurar hogar” que lleve a `/onboarding/household` (pero dejando claro que es un wizard).
  - [ ] Si SÍ hay API de update:
    - [ ] implementar `updateHousehold(...)` en `frontend/src/api.js` y permitir edición.

- [ ] **T006 — Enrutar HouseholdSettingsPage en AppLayout**
  - [ ] Añadir ruta `/settings/household` en `frontend/src/router.jsx`.

- [ ] **T007 — Eliminar “dead-end” de OnboardingFixedExpensesPage**
  - [ ] En `frontend/src/pages/OnboardingFixedExpensesPage.jsx`:
    - [ ] añadir botón **Volver a ajustes** → `/rules` cuando NO es onboarding (cuando `!isOnboarding`).
    - [ ] cuando SÍ es onboarding, añadir “Volver” → `/onboarding/cards`.
    - [ ] renombrar “Saltar por ahora” a un texto con intención clara (ej. “Ir al dashboard”).

- [ ] **T008 — Ajustes de estilos mínimos (si hace falta)**
  - [ ] Evitar uso de `.onboardingCard/.onboardingHeader` en pantallas nuevas de gestión.
  - [ ] Reusar estilos existentes; añadir CSS nuevo solo si es imprescindible.

- [ ] **T009 — Validación**
  - [ ] `cd frontend && npm run lint`
  - [ ] `cd frontend && npm run test:coverage -- --run`
  - [ ] Smoke manual (cuando se ejecute el frontend):
    - [ ] `/rules` → “Tarjetas” → volver a `/rules`
    - [ ] `/rules` → “Hogar” → volver a `/rules`
    - [ ] `/onboarding/fixed-expenses` tiene “Volver”

---

## Done Criteria
- Desde Settings (`/rules`) ya no se navega a rutas `/onboarding/*`.
- Las pantallas de gestión (`/cards`, `/settings/household`) no tienen visual onboarding.
- No existen flujos donde la única salida sea “Saltar por ahora”.
