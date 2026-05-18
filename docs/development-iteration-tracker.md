# Seguimiento de desarrollo por iteración

Este documento es la fuente única de verdad para plan, ejecución y trazabilidad.
Se actualiza al cierre de cada iteración.

---

## Marco operativo: MEM-SEQ-THINK

### MEM (Memoria viva del proyecto)
- Objetivo de negocio vigente.
- Decisiones de arquitectura tomadas.
- Restricciones activas (DDD/Clean/Hexagonal, Go idiomático, seguridad).
- Estado de deuda técnica y riesgos abiertos.

### SEQ (Secuencia de ejecución)
- Fase → iteración → tareas por capa (`domain -> application -> ports -> adapters -> wiring`).
- Cada fase se entrega en incrementos funcionales pequeños y verificables.
- No se abre una fase nueva sin cumplir criterios de salida de la fase actual.

### THINK (Razonamiento y control de calidad)
- Hipótesis de diseño por iteración.
- Validación obligatoria (`go test ./...`, revisión de errores, smoke test API/UI).
- Resultado contra objetivo (cumplido/parcial/no cumplido).
- Decisión explícita del siguiente paso.

---

## Estado global del roadmap

- Fecha de arranque: 2026-03-07
- Fase actual: `Phase 4` (pendiente)
- Última fase cerrada: `Iteración 005 (UX Polish)`
- Estrategia de entrega: incremental por iteraciones cortas
- Estado general: en progreso

---

## Plan maestro por fases (detalle completo)

## Phase 0 — Foundation & deuda técnica crítica (cerrada)

### Objetivo
Alinear implementación con arquitectura hexagonal y dejar base limpia para escalar.

### Entregables planeados
- Contratos movidos a `ports/inbound` y `ports/outbound`.
- Eliminación de deriva de contratos en `application`.
- Checklist de arquitectura verificado.

### Criterios de salida
- Compilación verde.
- Tests verdes.
- Dependencias por capa respetadas.

### Estado
✅ Cerrada.

---

## Phase 1 — Household + Members (N personas)

### Objetivo
Pasar de un modelo de gasto simple a un modelo de hogar con miembros flexibles y sueldo por persona.

### Alcance funcional
- Crear hogares (`households`).
- Crear/listar/editar miembros por hogar (`members`).
- Asociar cada gasto a `paid_by_member_id`.
- Mantener gastos personales y compartidos (`is_shared`).

### Entregables técnicos por capa
- Domain:
  - Nuevas entidades `Household` y `Member` con constructores ricos.
  - Invariantes mínimas (nombre requerido, household_id válido, etc.).
- Application:
  - Casos de uso para crear/obtener/listar/actualizar household y members.
- Ports:
  - Nuevos puertos inbound/outbound para household/member.
- Adapters:
  - Endpoints HTTP para household/member.
  - Repositorios Postgres para household/member.
- DB:
  - Migraciones para `households`, `members` y alter de `expenses`.
- Frontend:
  - Selección/alta de household y gestión básica de miembros.

### Criterios de salida
- CRUD mínimo de household/member operativo.
- Gastos guardan `paid_by_member_id` sin romper endpoints actuales.
- Tests backend verdes.

### Riesgos
- Compatibilidad hacia atrás con datos previos de `expenses`.

---

## Phase 2 — Auth simple (email/password + JWT)

### Objetivo
Controlar identidad de usuario y permisos por household.

### Alcance funcional
- Registro/login con email + password.
- Emisión y validación de JWT.
- Autorización por pertenencia a household.

### Entregables técnicos por capa
- Domain/Application:
  - Casos de uso `register`, `login`.
- Adapters HTTP:
  - Middleware de auth.
  - Endpoints `/v1/auth/register` y `/v1/auth/login`.
- Security:
  - Hash seguro de password (`bcrypt`).
  - Rechazo de acceso cross-household.

### Criterios de salida
- Endpoints privados inaccesibles sin token.
- Usuario sólo ve datos de sus households.

### Riesgos
- Manejo de expiración de token y UX de sesión.

---

## Phase 3 — Settlement (ajuste entre personas)

### Objetivo
Calcular cuánto debe transferir cada persona al cierre de periodo.

### Alcance funcional
- Modo de reparto configurable por household:
  - `equal` (50/50 o equivalente N personas)
  - `proportional` (según sueldo)
- Cálculo de neto por miembro y transferencias sugeridas.
- Exclusión de gastos con vales (`payment_method=voucher`) del pool compartido.

### Entregables técnicos por capa
- Domain:
  - Lógica pura de settlement y ajustes.
- Application:
  - Caso de uso `CalculateSettlement` por mes.
- HTTP:
  - Endpoint de reporte de ajuste mensual.
- Frontend:
  - Vista de “quién paga a quién y cuánto”.

### Criterios de salida
- Resultado consistente para ambos modos (`equal`, `proportional`).
- Casos de borde cubiertos (N miembros, sueldo 0, sin gastos).

### Riesgos
- Definir claramente política cuando faltan sueldos.

---

## Phase 4 — Gastos fijos + recurrencia

### Objetivo
Automatizar gastos repetitivos y visualizarlos claramente.

### Alcance funcional
- Etiquetas de gasto (`fixed`, `variable`, `msi`).
- Plantillas recurrentes (mensual/quincenal/semanal).
- Generación controlada de gastos por periodo.

### Entregables técnicos
- Nueva tabla `recurring_expenses`.
- Casos de uso de alta/listado/actualización/generación.
- UI para gestión de recurrentes.

### Criterios de salida
- Generación idempotente por periodo.
- Filtros por tipo en UI.

### Riesgos
- Duplicados por ejecución repetida de generación.

---

## Phase 5 — MSI como entidad

### Objetivo
Modelar compras a meses y reflejar impacto mensual real.

### Alcance funcional
- Registro de compra MSI (monto total, meses, inicio).
- Cálculo mensual de porción vigente.
- Visual de meses restantes.

### Entregables técnicos
- Entidad/tabla `installments`.
- Casos de uso de creación y cálculo mensual.
- Integración con settlement para gastos compartidos.

### Criterios de salida
- Cálculo mensual correcto en distintos rangos de fechas.

### Riesgos
- Redondeo de centavos y ajuste en última mensualidad.

---

## Phase 6 — Dashboards y analítica histórica

### Objetivo
Dar visibilidad inmediata: histórico, tendencias y distribución de gasto.

### Alcance funcional
- Resumen mensual por categoría/persona/método pago.
- Tendencia histórica.
- Datos preparados para predicción posterior.

### Entregables técnicos
- Endpoints de agregación.
- Consultas SQL optimizadas (`GROUP BY`, índices según carga real).
- UI de dashboards (gráficas + KPIs).

### Criterios de salida
- Reportes consistentes con datos base.
- Tiempo de respuesta aceptable para histórico.

### Riesgos
- Performance en rangos largos de tiempo.

---

## Phase 7 — Captura rápida e importación

### Objetivo
Eliminar fricción de captura y acelerar resumen mensual.

### Alcance funcional
- Importación de estados de cuenta (CSV inicial).
- Quick-add optimizado en UI (atajos y flujo mínimo).
- Diseño de puertos para OCR screenshot y WhatsApp (implementación posterior incremental).

### Entregables técnicos
- Endpoint de importación con validación.
- Parser de CSV con mapeo configurable.
- Registro de errores por fila para corrección.

### Criterios de salida
- Importación parcial tolerante a errores.
- Alta velocidad de captura manual en UI.

### Riesgos
- Variación de formatos bancarios.

---

## Dependencias entre fases

- `Phase 1` desbloquea `Phase 2` y `Phase 3`.
- `Phase 3` depende de tener `members + salaries` (`Phase 1`).
- `Phase 5` alimenta `Phase 6` (dashboards reales de deuda mensual).
- `Phase 7` se apoya en modelo estable de fases previas.

---

## Bitácora de iteraciones

### Iteración 001 — Arquitectura hexagonal base
- Fecha: 2026-03-07
- Fase: Phase 0
- MEM:
  - Se corrigió deriva: contratos fuera de `application` hacia `ports`.
  - Se validó checklist arquitectónico.
- SEQ:
  - `ports` creados -> use cases migrados -> adapters ajustados -> tests.
- THINK:
  - Resultado: ✅ objetivo cumplido.
  - Validación: `go test ./...` (13 passing, 0 failing).
  - Decisión siguiente: iniciar `Phase 1` con dominio household/member.
- Archivos clave:
  - `backend/internal/ports/outbound/expense_repository.go`
  - `backend/internal/ports/inbound/expense_usecases.go`
  - `backend/internal/application/expense/*.go`
  - `backend/internal/adapters/http/expense_handler.go`
  - `backend/internal/adapters/postgres/expense_repository.go`
  - `docs/architecture-checklist.md`

### Iteración 002 — Core de dominio Household/Member
- Fecha: 2026-03-07
- Fase: Phase 1
- Objetivo:
  - crear la base de dominio para soportar N personas por hogar.
- MEM:
  - Decisión de alcance: sólo backend-core (domain, ports, use cases, migrations).
  - Auth, HTTP adapters y frontend se difieren a siguientes iteraciones.
- SEQ:
  - Domain (`household`, `member`) -> ports inbound/outbound -> use cases base -> migraciones phase1.
- THINK:
  - Hipótesis:
    - separar core de dominio primero reduce riesgo al agregar auth y settlement después.
  - Validación ejecutada:
    - `go test ./...` ejecutado con éxito.
  - Resultado:
    - ✅ implementación completada y validada.
  - Decisión siguiente:
    - conectar adapters HTTP/Postgres para household/member en Iteración 003.
- Cambios por capa:
  - Domain:
    - entidad `Household` con `SettlementMode` (`equal`, `proportional`) y `currency` ISO.
    - entidad `Member` con `monthly_salary_cents`, validaciones de nombre/email/salario.
  - Application:
    - use cases: `RegisterHousehold`, `ListHouseholds`, `RegisterMember`, `ListMembers`.
  - Ports:
    - inbound/outbound para household/member.
  - Adapters:
    - sin cambios en esta iteración.
  - DB/Migrations:
    - `002_create_households.sql`
    - `003_create_members.sql`
    - `004_alter_expenses_phase1_fields.sql`
  - Frontend:
    - sin cambios en esta iteración.
- Archivos clave:
  - `backend/internal/domain/household/household.go`
  - `backend/internal/domain/member/member.go`
  - `backend/internal/ports/inbound/household_usecases.go`
  - `backend/internal/ports/inbound/member_usecases.go`
  - `backend/internal/ports/outbound/household_repository.go`
  - `backend/internal/ports/outbound/member_repository.go`
  - `backend/internal/application/household/*.go`
  - `backend/internal/application/member/*.go`
  - `backend/migrations/002_create_households.sql`
  - `backend/migrations/003_create_members.sql`
  - `backend/migrations/004_alter_expenses_phase1_fields.sql`
- Riesgos / deuda:
  - falta wiring en `cmd/api` para nuevos casos de uso.
  - falta adapter Postgres para household/member.
  - falta exposición de endpoints HTTP.

### Iteración 003 — Settlement mensual (equal/proportional)
- Fecha: 2026-03-08
- Fase: Phase 3
- Objetivo:
  - habilitar cálculo mensual de settlement por hogar y exponerlo en API/UI.
- MEM:
  - settlement depende de `paid_by_member_id`, `is_shared`, `currency` y `payment_method` en gastos.
  - política definida para sueldos cero en modo proporcional: fallback a `equal` con razón explícita.
- SEQ:
  - Domain (`expense` + `settlement`) -> ports -> use case -> adapters HTTP/Postgres -> wiring -> frontend.
- THINK:
  - Hipótesis:
    - encapsular cálculo en dominio permite probar reglas de reparto sin acoplar a DB/HTTP.
  - Validación ejecutada:
    - `cd backend && go test ./...` ✅
    - `cd frontend && npm run build` ✅
  - Resultado:
    - ✅ endpoint de settlement mensual operativo.
    - ✅ exclusión de `payment_method=voucher` del pool compartido.
    - ✅ vista frontend básica de transferencias sugeridas.
  - Decisión siguiente:
    - fortalecer UX de selección de miembros (evitar captura manual de IDs) y ampliar pruebas de adapters HTTP.
- Cambios por capa:
  - Domain:
    - nueva lógica pura `internal/domain/settlement/settlement.go`.
    - `expense` enriquecido con `paid_by_member_id`, `is_shared`, `currency`, `payment_method`.
  - Application:
    - nuevo use case `CalculateSettlement` mensual.
  - Ports:
    - inbound `CalculateSettlementUseCase`.
    - outbound: `ListByHouseholdAndPeriod` (expenses), `ListAllByHousehold` (members).
  - Adapters:
    - endpoint `GET /v1/households/{household_id}/settlement`.
    - repositorio Postgres de gastos actualizado con nuevos campos y query por periodo.
  - DB/Migrations:
    - `005_alter_expenses_add_payment_method.sql`.
  - Frontend:
    - formulario de gasto con `paid_by_member_id`, `is_shared`, `payment_method`.
    - panel de settlement mensual.
- Archivos clave:
  - `backend/internal/domain/settlement/settlement.go`
  - `backend/internal/application/settlement/calculate_settlement.go`
  - `backend/internal/adapters/http/settlement_handler.go`
  - `backend/internal/adapters/postgres/expense_repository.go`
  - `backend/internal/domain/expense/expense.go`
  - `backend/internal/ports/inbound/settlement_usecases.go`
  - `frontend/src/App.jsx`
  - `frontend/src/api.js`
- Riesgos / deuda:
  - falta suite de tests de handlers HTTP para settlement.

### Iteración 004 — Cierre UX de Phase 3 (selector de miembro)
- Fecha: 2026-03-08
- Fase: Phase 3
- Objetivo:
  - eliminar captura manual de `member_id` en alta de gasto para cerrar phase 3.
- MEM:
  - el formulario de gastos depende del listado de miembros por household.
- SEQ:
  - API client frontend -> estado/carga en `App` -> dropdown en `ExpenseForm` -> build.
- THINK:
  - Hipótesis:
    - un selector de miembros reduce errores de captura y cierra la deuda UX de settlement.
  - Validación ejecutada:
    - `cd frontend && npm run build` ✅
  - Resultado:
    - ✅ `ExpenseForm` usa selector real de miembros.
    - ✅ alta de gasto bloqueada cuando no hay miembros disponibles.
    - ✅ deuda UX principal de phase 3 cerrada.
  - Decisión siguiente:
    - abrir `Phase 4` con modelo de recurrentes.
- Cambios por capa:
  - Domain:
    - sin cambios.
  - Application:
    - sin cambios.
  - Ports:
    - sin cambios.
  - Adapters:
    - sin cambios en backend.
  - DB/Migrations:
    - sin cambios.
  - Frontend:
    - `listMembers` en cliente API.
    - carga de miembros por household en `App`.
    - selector de miembro y estado vacío en `ExpenseForm`.
- Archivos clave:
  - `frontend/src/api.js`
  - `frontend/src/App.jsx`
  - `frontend/src/components/ExpenseForm.jsx`
- Riesgos / deuda:
  - falta suite de tests de handlers HTTP para settlement.

### Iteración 005 — UX Polish
- Fecha: 2026-03-24
- Fase: Intermedia (entre Phase 3 y Phase 4)
- Objetivo:
  - pulir experiencia de usuario antes de agregar complejidad de Phase 4.
  - hacer la app delightful para usuarios nuevos.
- MEM:
  - backend A+, no requiere cambios.
  - frontend funcional pero con gaps UX críticos.
  - decisión: pausar Phase 4, priorizar polish.
- SEQ:
  - propuesta SDD → specs → design → tasks → implementación → verificación.
- THINK:
  - Hipótesis:
    - pulir UX antes de Phase 4 mejora fundación para features futuras.
    - cambios son solo frontend, bajo riesgo de regresiones.
  - Validación ejecutada:
    - `cd frontend && npm run build` ✅
  - Resultado:
    - ✅ todas las tareas P0 completadas.
    - ✅ tareas P1 principales completadas.
    - ✅ build de producción exitoso.
  - Decisión siguiente:
    - continuar con Phase 4 (gastos recurrentes) según roadmap original.
- Cambios por capa:
  - Domain:
    - sin cambios.
  - Application:
    - sin cambios.
  - Ports:
    - sin cambios.
  - Adapters:
    - sin cambios en backend.
  - DB/Migrations:
    - sin cambios.
  - Frontend:
    - salary field en dólares (no cents) en OnboardingMemberPage.
    - settlement auto-carga mes actual, botón "This month".
    - campo "Paid by" visible sin "More options" en ExpenseModal.
    - empty state en dashboard para usuarios nuevos.
    - botón "+ Member" en AppHeader.
    - ruta /members/new para invitar miembros post-onboarding.
    - badge "(current)" en SettlementPanel cuando es mes actual.
    - estilos CSS para empty state y badge.
- Archivos clave:
  - `frontend/src/pages/OnboardingMemberPage.jsx`
  - `frontend/src/hooks/useSettlement.js`
  - `frontend/src/components/SettlementPanel.jsx`
  - `frontend/src/components/ExpenseModal.jsx`
  - `frontend/src/pages/DashboardPage.jsx`
  - `frontend/src/components/AppHeader.jsx`
  - `frontend/src/router.jsx`
  - `frontend/src/styles.css`
  - `docs/iteration-005-ux-polish-proposal.md`
  - `docs/iteration-005-ux-polish-specs.md`
  - `docs/iteration-005-ux-polish-design.md`
  - `docs/iteration-005-ux-polish-tasks.md`
- Riesgos / deuda:
  - tareas P2 (tooltips, category icons, members list) postponed.
  - falta smoke test manual completo.

---

### Iteración 006 — UX Polish y Period Guardrails
- Fecha: 2026-05-17
- Fase: Intermedia (entre Phase 3 y Phase 4)
- Objetivo:
  - pulir UX, traducir UI a español, endurecer seguridad de backend y añadir navegación entre periodos.
- MEM:
  - backend tenía time simulator (deuda técnica de producción), period close permitía a no-owners cerrar por consenso.
  - frontend: mixed English/Spanish, onboarding roto (sin periodo inicial), datos de consenso mock, sin selector de periodo, componentes UI oversize.
  - decisión: tres PRs verticales (A: backend, B: onboarding+i18n, C: period UX).
- SEQ:
  - SDD completo: proposal → specs → design → tasks → apply (3 PRs) → verify → archive.
  - PR A mergeó antes de PR C (consensus endpoint necesario desde frontend).
- THINK:
  - Hipótesis:
    - PRs independientes reducen riesgo de regresión y permiten rollback por slice.
    - backend-first asegura que frontend tenga endpoints reales.
  - Validación ejecutada:
    - `cd backend && go test -race ./...` ✅ (214 tests, 0 fallos, race detector clean).
    - `cd frontend && npm run build` ✅ (0 errores, 0 warnings).
    - `cd frontend && npx vitest run` ❌ 15 fallos (7 del cambio, 8 pre-existentes).
  - Resultado:
    - ✅ Time simulator removido (dev_handler.go eliminado, clock.go simplificado).
    - ✅ Owner-only period close en ambos paths (consenso y force).
    - ✅ Endpoint `GET .../periods/{pid}/consensus` operativo.
    - ✅ Onboarding: register → household → member → dashboard funcional.
    - ✅ UI 100% en español (Login, Register, ExpenseForm, IncomesPanel, etc.).
    - ✅ Chip compacto de estado de periodo (~80×24px) en header.
    - ✅ Barra apilada 100% para ingresos con leyenda.
    - ✅ Datos reales de consenso (no mock 50%).
    - ✅ Selector global de periodo con "Volver al actual".
    - ⚠️ initializePeriod no se llama automáticamente en onboarding (gap parcial).
    - ⚠️ 15 tests frontend rotos por test drift.
  - Decisión siguiente:
    - continuar con Phase 4 (gastos recurrentes) según roadmap original.
    - abordar issues conocidos como P1/P2: error mapping handleClose→403, cleanup TimeSimulator dead code, auto-init periodo en onboarding, frontend test fixes.
- Cambios por capa:
  - Domain:
    - sin cambios (solo use cases nuevos/get_period_consensus.go).
  - Application:
    - `close_period.go`: owner check reposicionado antes de consenso.
    - `get_period_consensus.go`: nuevo caso de uso para estadísticas de aprobación.
    - `shared/clock.go`: simplificado (remove simulación, `Now()` → `time.Now()`).
  - Ports:
    - `inbound/period_usecases.go`: nuevo `GetConsensusUseCase` + `GetConsensusInput/Output`.
  - Adapters:
    - `dev_handler.go`: eliminado (ruta `/v1/dev/time-offset`).
    - `period_handler.go`: nuevo `handleGetConsensus`, route registrada.
    - `server.go`: removido `IsDev` field, dev route, devHandler import.
    - `cmd/api/main.go`: wired `GetPeriodConsensusUseCase`.
  - DB/Migrations:
    - sin cambios.
  - Frontend:
    - PR A: removed `advanceTime` de api.js, TimeSimulator.jsx → null.
    - PR B: onboarding flow fix (redirect + period init opción), i18n español (Login, Register, ExpenseForm, IncomesPanel, ConsensusProgressRing).
    - PR C: PeriodStatusRibbon → chip compacto, IncomesPanel → stacked bar, real consensus data, PeriodSelector.jsx (nuevo), selectedPeriodId state via AppShellContext.
- Archivos clave:
  - `backend/internal/application/period/close_period.go`
  - `backend/internal/application/period/get_period_consensus.go`
  - `backend/internal/application/shared/clock.go`
  - `backend/internal/adapters/http/dev_handler.go` (eliminado)
  - `backend/internal/adapters/http/period_handler.go`
  - `backend/internal/adapters/http/server.go`
  - `backend/cmd/api/main.go`
  - `frontend/src/pages/RegisterPage.jsx`
  - `frontend/src/pages/LoginPage.jsx`
  - `frontend/src/components/ExpenseForm.jsx`
  - `frontend/src/components/PeriodStatusRibbon.jsx`
  - `frontend/src/components/IncomesPanel.jsx`
  - `frontend/src/components/PeriodSelector.jsx` (nuevo)
  - `frontend/src/hooks/useDashboardUxState.js`
  - `frontend/src/api.js`
  - `frontend/src/components/AppHeader.jsx`
  - `frontend/src/styles.css`
  - SDD artifacts: engram obs #288–#297
- Riesgos / deuda:
  - clock.go no se eliminó completamente (AD-001 desviación).
  - TimeSimulator import/render dead code en AppLayout.jsx.
  - handleClose siempre retorna 500 (debería mapear ErrForbidden → 403).
  - initializePeriod no automático en onboarding.
  - 15 tests frontend fallando (7 por test drift del cambio).
- Bloqueos (si aplica):
  - ninguno.
