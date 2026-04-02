# Iteración 005 — UX Polish — Specifications

**Propuesta**: `iteration-005-ux-polish-proposal.md`  
**Estado**: Especificación  
**Fecha**: 2026-03-24  
**Última revisión**: 2026-04-01

---

## Implementation Status (2026-04-01)

| ID | User Story | Status | Evidence |
|----|------------|--------|----------|
| US-001 | Onboarding salary in dollars | ✅ DONE | `OnboardingMemberPage.jsx` line 89 uses `$` prefix |
| US-002 | Settlement auto-select current month | ✅ DONE | `useSettlement.js` lines 13-14 init with current date |
| US-003 | "Invite Member" button | ✅ DONE | `AppHeader.jsx` lines 44-53 shows "+ Member" link |
| US-004 | "Paid by" field always visible | ✅ DONE | `ExpenseModal.jsx` has visible `Paid by` `<select>` outside advanced block |
| US-005 | Dashboard empty state | ✅ DONE | `DashboardPage.jsx` lines 179-187 |
| US-006 | Tooltips on key fields | ✅ DONE | `Tooltip.jsx` + usage in `ExpenseModal.jsx` and `SettlementPanel.jsx` |
| US-007 | Visual polish (icons, colors) | ⚠️ PARTIAL | Icons/animations in place; color refinement still open |
| US-008 | Members list visible | ✅ DONE | `MembersPanel.jsx` rendered from `DashboardPage.jsx` |

### Remaining Work

1. **US-007**: Final color token refinement and visual QA pass
2. **Verification**: Run full smoke checklist and mobile viewport validation

---

## Overview

Este documento especifica los requisitos funcionales y de UX para la Iteración 005, organizado por user stories con escenarios Given/When/Then.

---

## US-001: Onboarding con salario en dólares

### User Story
**As a** new user setting up my member profile  
**I want to** enter my monthly salary in dollars (not cents)  
**So that** I don't get confused by the unit and enter correct data

### Current Behavior (Problem)
- Label dice "Monthly salary (cents, optional)"
- Usuario debe convertir manualmente $5,000 → 500,000 cents
- Alta probabilidad de error (olvidar multiplicar por 100)

### Desired Behavior (Solution)
- Label dice "Monthly salary (optional)"
- Campo acepta decimales (e.g., "5000" o "5000.50")
- Frontend convierte a cents antes de enviar al backend

### Acceptance Criteria

#### Scenario 1: Entering salary in dollars
```gherkin
Given I am on the "Add a member" onboarding page
When I see the salary field
Then the label MUST say "Monthly salary (optional)"
And the placeholder SHOULD say "e.g., 5000"
And the input type SHOULD be "number" with step="0.01"
```

#### Scenario 2: Submitting salary
```gherkin
Given I entered "5000" in the salary field
When I submit the form
Then the frontend MUST convert "5000" to 500000 cents
And the backend MUST receive `monthlySalaryCents: 500000`
```

#### Scenario 3: Decimal input
```gherkin
Given I entered "5000.50" in the salary field
When I submit the form
Then the frontend MUST convert "5000.50" to 500050 cents
```

### Files Affected
- `frontend/src/pages/OnboardingMemberPage.jsx` (lines 87-98)

---

## US-002: Settlement auto-selecciona mes actual

### User Story
**As a** user viewing my dashboard  
**I want** the settlement panel to automatically show the current month  
**So that** I don't have to manually select year and month every time

### Current Behavior (Problem)
- Settlement panel carga con un mes/año arbitrario
- Usuario debe seleccionar año + mes + click "Refresh"
- 3 acciones para ver datos relevantes

### Desired Behavior (Solution)
- Al cargar dashboard, settlement automáticamente muestra mes/año actual
- Botón "This month" para resetear rápidamente a mes actual

### Acceptance Criteria

#### Scenario 1: Auto-load current month on mount
```gherkin
Given I just logged in and landed on the dashboard
When the settlement panel renders for the first time
Then the year selector MUST show the current year (e.g., 2026)
And the month selector MUST show the current month (e.g., 3 for March)
And the settlement data MUST auto-load without clicking "Refresh"
```

#### Scenario 2: "This month" button
```gherkin
Given I changed the settlement period to "January 2025"
And I am viewing the settlement panel
When I click the "This month" button
Then the year/month MUST reset to current date
And the settlement data MUST reload automatically
```

#### Scenario 3: Visual indicator of current period
```gherkin
Given the settlement panel is showing the current month
When I look at the panel header
Then I SHOULD see a visual indicator like "March 2026 (current)"
```

### Files Affected
- `frontend/src/hooks/useSettlement.js` (inicializar con fecha actual)
- `frontend/src/components/SettlementPanel.jsx` (agregar botón "This month", visual indicator)

---

## US-003: Botón "Invite Member"

### User Story
**As a** user who completed onboarding  
**I want** a button to invite additional members  
**So that** I can add roommates/family after initial setup

### Current Behavior (Problem)
- No hay forma de agregar miembros después del onboarding
- Usuario debe adivinar una URL o buscar en código
- El endpoint existe, pero no hay UI

### Desired Behavior (Solution)
- Botón "Invite Member" visible en app header o dashboard
- Al hacer click, navega a `/members/new` o abre modal
- Reutiliza `OnboardingMemberPage` o crea componente similar

### Acceptance Criteria

#### Scenario 1: Button visibility
```gherkin
Given I am logged in and viewing the dashboard
When I look at the app header or top section
Then I MUST see a button labeled "Invite Member" or "+ Member"
```

#### Scenario 2: Navigation to member form
```gherkin
Given I am on the dashboard
When I click "Invite Member"
Then I SHOULD navigate to "/members/new"
Or a modal SHOULD open with the member form
```

#### Scenario 3: Form reuses onboarding logic
```gherkin
Given I am on the "Invite Member" page/modal
When I submit the form
Then the behavior MUST be identical to onboarding member creation
And I MUST be redirected back to dashboard after success
```

### Files Affected
- `frontend/src/components/AppHeader.jsx` (agregar botón)
- `frontend/src/router.jsx` (agregar ruta `/members/new`)
- `frontend/src/pages/OnboardingMemberPage.jsx` (reutilizar o refactorizar en componente compartido)

---

## US-004: Campo "Paid by" siempre visible

### User Story
**As a** user adding an expense  
**I want to** see "who paid" without expanding "More options"  
**So that** I can quickly verify or change the payer

### Current Behavior (Problem)
- "Paid by" está oculto en sección "More options" (collapsed por defecto)
- Usuario no ve que el payer está auto-seleccionado
- Puede causar errores (gasto asignado a persona incorrecta)

### Desired Behavior (Solution)
- "Paid by" visible siempre, entre "Description" y "Category"
- Dropdown con lista de miembros
- Hint: "Defaults to you" si current member está seleccionado

### Acceptance Criteria

#### Scenario 1: Field always visible
```gherkin
Given I open the "New expense" modal
When the form renders
Then I MUST see a "Paid by" field immediately
And it MUST be a dropdown with member names
And it MUST NOT be inside the "More options" section
```

#### Scenario 2: Default selection
```gherkin
Given my linked member is "Ale"
When I open the "New expense" modal
Then the "Paid by" dropdown MUST default to "Ale"
And there SHOULD be a hint: "Defaults to you"
```

#### Scenario 3: Change payer
```gherkin
Given the "Paid by" field defaults to "Ale"
When I change it to "Neba"
And I submit the form
Then the expense MUST be saved with `paidByMemberId` = Neba's ID
```

### Files Affected
- `frontend/src/components/ExpenseModal.jsx` (mover campo fuera de `showAdvanced`)

---

## US-005: Empty state en dashboard

### User Story
**As a** new user with no expenses  
**I want to** see a helpful empty state  
**So that** I understand what to do next

### Current Behavior (Problem)
- Dashboard muestra paneles vacíos sin contexto
- Usuario no sabe qué hacer (no hay CTA)
- Sensación de app "rota" o incompleta

### Desired Behavior (Solution)
- Si no hay gastos, mostrar empty state con:
  - Icono/ilustración
  - Título: "No expenses yet"
  - Mensaje: "Tap the + button below to add your first expense"
  - Flecha apuntando al FAB (opcional)

### Acceptance Criteria

#### Scenario 1: Empty state when no expenses
```gherkin
Given I am a new user with no expenses recorded
When I view the dashboard
Then I MUST see an empty state message
And the message MUST guide me to click the "+" button
```

#### Scenario 2: Empty state disappears after adding expense
```gherkin
Given I see the empty state
When I add my first expense
Then the empty state MUST disappear
And the expense list MUST show my new expense
```

#### Scenario 3: Empty state styling
```gherkin
Given the empty state is visible
Then it MUST be centered in the viewport
And it SHOULD have a subtle background or border
And it SHOULD use friendly, encouraging copy
```

### Files Affected
- `frontend/src/pages/DashboardPage.jsx` (agregar conditional render para empty state)
- `frontend/src/styles.css` (estilos para empty state)

---

## US-006: Tooltips explicativos

### User Story
**As a** user unfamiliar with the app  
**I want** tooltips on key fields and panels  
**So that** I understand what each section does

### Current Behavior (Problem)
- Ningún tooltip o help text
- Usuario debe adivinar qué significa "Settlement mode", "Shared expense", etc.

### Desired Behavior (Solution)
- Agregar icono "?" junto a labels complejos
- Hover/click muestra tooltip con explicación breve
- Implementación simple (CSS `title` attribute o componente `<Tooltip>`)

### Acceptance Criteria

#### Scenario 1: Tooltip on "Shared expense"
```gherkin
Given I am filling the expense form
When I hover over the "Shared expense" checkbox label
Then I SHOULD see a tooltip: "Shared expenses are split among all members"
```

#### Scenario 2: Tooltip on "Settlement mode"
```gherkin
Given I am viewing the settlement panel
When I hover over "Mode: proportional"
Then I SHOULD see a tooltip: "Expenses are split based on each member's salary"
```

#### Scenario 3: Tooltip implementation
```gherkin
Given tooltips are implemented
Then they MUST use either HTML `title` attribute
Or a custom `<Tooltip>` component with accessible ARIA labels
```

### Files Affected
- `frontend/src/components/ExpenseModal.jsx` (tooltip en "Shared expense")
- `frontend/src/components/SettlementPanel.jsx` (tooltip en "Mode")
- `frontend/src/ui/Tooltip.jsx` (nuevo componente opcional)

---

## US-007: Visual polish (colores, iconos, spacing)

### User Story
**As a** user  
**I want** the app to feel polished and professional  
**So that** I trust it with my financial data

### Current Behavior (Problem)
- Colores genéricos
- Sin iconos en categorías
- Spacing inconsistente
- Transiciones abruptas

### Desired Behavior (Solution)
- Agregar iconos a categorías (🏠 Rent, 🚗 Auto, 🍔 Food, etc.)
- Mejorar colores de estados (green/red/yellow más vibrantes)
- Consistentizar spacing (usar tokens CSS)
- Suavizar animaciones de modales y toasts

### Acceptance Criteria

#### Scenario 1: Category icons
```gherkin
Given I am viewing an expense with category "rent"
When I see the expense card
Then I SHOULD see a house icon (🏠) next to "Rent"
```

#### Scenario 2: Color palette
```gherkin
Given the app uses color coding
Then success states MUST use a vibrant green (#22c55e)
And error states MUST use a vibrant red (#ef4444)
And warning states MUST use a vibrant yellow/orange (#f59e0b)
```

#### Scenario 3: Smooth animations
```gherkin
Given I open a modal
Then the modal MUST fade in smoothly (200-300ms)
And the backdrop MUST have a subtle blur effect
```

#### Scenario 4: Consistent spacing
```gherkin
Given the app uses CSS spacing
Then all spacing MUST use tokens (e.g., --space-1, --space-2, etc.)
And components MUST have consistent padding/margins
```

### Files Affected
- `frontend/src/styles.css` (color tokens, spacing tokens, animations)
- `frontend/src/components/ExpenseModal.jsx` (category icons)
- `frontend/src/ui/Banner.jsx` (color updates)

---

## US-008: Lista de miembros visible

### User Story
**As a** household admin  
**I want to** see all members in my household  
**So that** I know who has access and can manage them

### Current Behavior (Problem)
- No hay vista de "Members" en el frontend
- Usuario no puede ver quién está en su household
- No puede verificar permisos o remover miembros

### Desired Behavior (Solution)
- Agregar sección "Members" en dashboard o página separada
- Mostrar lista de miembros con nombre, email, rol
- (Opcional) Botón "Remove" para admin

### Acceptance Criteria

#### Scenario 1: Members list in dashboard
```gherkin
Given I am viewing the dashboard
When I scroll to the members section
Then I MUST see a list of all household members
And each member MUST show: name, email, role
```

#### Scenario 2: Separate members page
```gherkin
Given I am on the dashboard
When I click "View Members" or navigate to "/members"
Then I SHOULD see a dedicated members page
And the page SHOULD list all members with details
```

#### Scenario 3: Empty state for members
```gherkin
Given my household has only me as a member
When I view the members list
Then I SHOULD see: "You're the only member. Invite others to split expenses!"
```

### Files Affected
- `frontend/src/pages/DashboardPage.jsx` (agregar sección de members) OR
- `frontend/src/pages/MembersPage.jsx` (nueva página)
- `frontend/src/components/MembersList.jsx` (nuevo componente)

---

## Non-Functional Requirements

### NFR-001: Performance
- Todas las mejoras UI MUST NOT aumentar el tiempo de carga inicial en >200ms
- Animaciones MUST usar CSS transforms (no layout changes) para 60fps

### NFR-002: Accessibility
- Todos los tooltips MUST tener ARIA labels apropiados
- Botones nuevos MUST tener labels descriptivos para screen readers
- Color contrast MUST cumplir WCAG AA (4.5:1 para texto normal)

### NFR-003: Responsiveness
- Todos los cambios MUST funcionar en mobile (viewport 375px)
- Modales MUST adaptar su tamaño en pantallas pequeñas

### NFR-004: Browser Compatibility
- Debe funcionar en: Chrome 90+, Firefox 88+, Safari 14+, Edge 90+

---

## Verification Plan

### Manual Testing Checklist

#### Onboarding
- [ ] Salary field muestra label "Monthly salary (optional)"
- [ ] Input acepta decimales
- [ ] Backend recibe cents correctamente

#### Settlement
- [ ] Auto-load mes/año actual al montar dashboard
- [ ] Botón "This month" funciona
- [ ] Visual indicator muestra "(current)" cuando aplica

#### Members
- [ ] Botón "Invite Member" visible y funcional
- [ ] Formulario funciona igual que onboarding
- [ ] Lista de miembros visible

#### Expense Form
- [ ] "Paid by" visible sin expandir "More options"
- [ ] Dropdown muestra miembros correctamente
- [ ] Default selection funciona

#### Empty States
- [ ] Empty state aparece cuando no hay gastos
- [ ] Desaparece después de agregar primer gasto
- [ ] Copy es claro y amigable

#### Visual Polish
- [ ] Categorías tienen iconos
- [ ] Colores vibrantes en success/error/warning
- [ ] Animaciones suaves
- [ ] Spacing consistente

### Smoke Test Flow
1. Registrar usuario nuevo
2. Crear household
3. Verificar empty state
4. Agregar primer miembro (verificar salary field)
5. Agregar primer gasto (verificar "Paid by" visible)
6. Verificar settlement auto-carga mes actual
7. Click botón "Invite Member", agregar segundo miembro
8. Verificar lista de miembros
9. Agregar segundo gasto
10. Verificar animaciones y colores

---

## Definition of Done

- [ ] Todas las user stories completadas
- [ ] Todos los acceptance criteria cumplidos
- [ ] Manual testing checklist 100% pasado
- [ ] Smoke test flow completado sin errores
- [ ] Build de frontend exitoso (`npm run build`)
- [ ] No errores de console en browser
- [ ] Funciona en mobile viewport
- [ ] Code review (self-review mínimo)
- [ ] Documentation updated (iteration tracker)

---

## Related Documents

- `docs/iteration-005-ux-polish-proposal.md` — Propuesta aprobada
- `docs/iteration-005-ux-polish-design.md` — Diseño técnico (siguiente paso)
- `docs/iteration-005-ux-polish-tasks.md` — Breakdown de tareas (siguiente paso)
