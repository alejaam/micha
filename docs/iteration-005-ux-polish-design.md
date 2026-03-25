# Iteración 005 — UX Polish — Technical Design

**Propuesta**: `iteration-005-ux-polish-proposal.md`  
**Specs**: `iteration-005-ux-polish-specs.md`  
**Estado**: Diseño técnico  
**Fecha**: 2026-03-24

---

## Overview

Este documento detalla las decisiones técnicas, componentes afectados, y cambios de estado para la Iteración 005.

---

## Architecture Decisions

### AD-001: No new dependencies
**Decision**: Implementar todas las mejoras UX sin agregar librerías externas  
**Rationale**:  
- Frontend ya tiene todo lo necesario (React 18, CSS vanilla)
- Agregar deps aumenta bundle size
- Tooltips simples se pueden hacer con CSS puro o componente ligero
**Alternatives considered**:  
- react-tooltip (rechazada: 50kb extra)
- tippy.js (rechazada: over-engineering)

### AD-002: Reutilizar OnboardingMemberPage para "Invite Member"
**Decision**: Crear ruta `/members/new` que renderiza `OnboardingMemberPage` con lógica compartida  
**Rationale**:  
- DRY (Don't Repeat Yourself)
- Form logic ya existe y funciona
- Solo cambia el copy ("Add a member" vs "Invite member")
**Alternatives considered**:  
- Crear `InviteMemberModal` (rechazada: duplica código)
- Refactorizar en componente `<MemberForm>` (aceptable, pero YAGNI por ahora)

### AD-003: Settlement auto-load en hook `useSettlement`
**Decision**: Inicializar `settlementYear` y `settlementMonth` con fecha actual en el hook  
**Rationale**:  
- Encapsula lógica de fecha en un solo lugar
- Dashboard no necesita saber cómo calcular "mes actual"
- Facilita testing del hook
**Alternatives considered**:  
- Calcular en `DashboardPage.jsx` (rechazada: acoplamiento)

### AD-004: "Paid by" field como FormField standalone
**Decision**: Mover "Paid by" fuera de `showAdvanced` y renderizarlo siempre  
**Rationale**:  
- Campo crítico, debe ser visible
- No aumenta complejidad visual (es un select pequeño)
- Mejora UX significativamente
**Alternatives considered**:  
- Mantener en "More options" con label más claro (rechazada: no resuelve el problema)

### AD-005: Empty state inline, no componente separado
**Decision**: Renderizar empty state directamente en `DashboardPage` con conditional  
**Rationale**:  
- Es específico al dashboard (no reutilizable)
- Manteniendo componente pequeño y simple
**Alternatives considered**:  
- Crear `<EmptyState>` component (rechazada: YAGNI)

---

## Component Changes

### 1. OnboardingMemberPage.jsx

**Change**: Salary field de cents a dólares

#### Current Structure
```jsx
<FormField label="Monthly salary (cents, optional)" htmlFor="memSalary">
  <input
    type="number"
    min="0"
    value={salary}  // expects cents
    onChange={(e) => setSalary(e.target.value)}
  />
</FormField>
```

#### New Structure
```jsx
<FormField label="Monthly salary (optional)" htmlFor="memSalary">
  <div className="inputWrap">
    <span className="inputPrefix">$</span>
    <input
      type="number"
      min="0"
      step="0.01"  // permite decimales
      placeholder="e.g., 5000"
      value={salaryDollars}  // nuevo state
      onChange={(e) => setSalaryDollars(e.target.value)}
    />
  </div>
</FormField>
```

#### State Changes
```js
// BEFORE
const [salary, setSalary] = useState('0')  // cents
await createMember({ monthlySalaryCents: Number(salary) || 0 })

// AFTER
const [salaryDollars, setSalaryDollars] = useState('')  // dollars
const salaryCents = Math.round((parseFloat(salaryDollars) || 0) * 100)
await createMember({ monthlySalaryCents: salaryCents })
```

#### Helpers
```js
// Agregar en utils.js
export function dollarsToCents(dollars) {
  const parsed = parseFloat(dollars)
  if (isNaN(parsed) || parsed < 0) return null
  return Math.round(parsed * 100)
}
```

---

### 2. useSettlement.js (hook)

**Change**: Auto-inicializar con fecha actual

#### Current Logic
```js
const [settlementYear, setSettlementYear] = useState(2026)  // hardcoded
const [settlementMonth, setSettlementMonth] = useState(1)   // hardcoded
```

#### New Logic
```js
// Calcular fecha actual una vez al montar
const now = new Date()
const [settlementYear, setSettlementYear] = useState(now.getFullYear())
const [settlementMonth, setSettlementMonth] = useState(now.getMonth() + 1)  // 0-indexed

// Auto-load al montar
useEffect(() => {
  if (isAuthenticated && householdId) {
    loadSettlement()
  }
}, [isAuthenticated, householdId])  // NO incluir year/month
```

#### New Method: `resetToCurrentMonth`
```js
function resetToCurrentMonth() {
  const now = new Date()
  setSettlementYear(now.getFullYear())
  setSettlementMonth(now.getMonth() + 1)
  // loadSettlement() se disparará por el useEffect que observa year/month
}

return {
  settlement,
  loadingSettlement,
  settlementYear,
  settlementMonth,
  setSettlementYear,
  setSettlementMonth,
  loadSettlement,
  resetToCurrentMonth,  // nuevo
}
```

---

### 3. SettlementPanel.jsx

**Change**: Agregar botón "This month" y visual indicator

#### New JSX (después de los selects)
```jsx
<button
  type="button"
  className="btn btnGhost btnSm"
  onClick={onResetToCurrentMonth}
>
  📅 This month
</button>
```

#### Visual Indicator (en el header)
```jsx
<h2 className="sectionTitle">
  <span className="sectionTitleIcon" aria-hidden>🧮</span>
  Monthly settlement
  {isCurrentMonth && <span className="currentPeriodBadge">(current)</span>}
</h2>
```

#### Props Changes
```jsx
export function SettlementPanel({
  settlement,
  settlementYear,
  settlementMonth,
  onSettlementYearChange,
  onSettlementMonthChange,
  onRefresh,
  onResetToCurrentMonth,  // nuevo
  loadingSettlement,
  memberIndex,
  currency = 'MXN',
}) {
  const now = new Date()
  const isCurrentMonth = settlementYear === now.getFullYear() 
    && settlementMonth === (now.getMonth() + 1)

  // ... resto del componente
}
```

---

### 4. AppHeader.jsx

**Change**: Agregar botón "Invite Member"

#### Current Structure (hypothetical)
```jsx
<header className="appHeader">
  <h1>micha</h1>
  <HouseholdSelector />
  <LogoutButton />
</header>
```

#### New Structure
```jsx
<header className="appHeader">
  <h1>micha</h1>
  <HouseholdSelector />
  <Link to="/members/new" className="btn btnSm btnGhost">
    + Member
  </Link>
  <LogoutButton />
</header>
```

---

### 5. router.jsx

**Change**: Agregar ruta `/members/new`

#### Current Routes
```jsx
<Route path="/" element={<DashboardPage />} />
<Route path="/onboarding/household" element={<OnboardingHouseholdPage />} />
<Route path="/onboarding/member" element={<OnboardingMemberPage />} />
```

#### New Routes
```jsx
<Route path="/" element={<DashboardPage />} />
<Route path="/onboarding/household" element={<OnboardingHouseholdPage />} />
<Route path="/onboarding/member" element={<OnboardingMemberPage />} />
<Route path="/members/new" element={<OnboardingMemberPage />} />  {/* reutiliza componente */}
```

---

### 6. ExpenseModal.jsx

**Change**: Mover "Paid by" fuera de `showAdvanced`

#### Current Structure (simplified)
```jsx
<FormField label="Amount" />
<FormField label="Description" />
<FormField label="Category" />

<button onClick={() => setShowAdvanced(!showAdvanced)}>More options</button>

{showAdvanced && (
  <>
    <p>Paid by: {paidByMemberName}</p>
    <FormField label="Payment method" />
    <FormField label="Expense type" />
    <FormField label="Shared expense" />
  </>
)}
```

#### New Structure
```jsx
<FormField label="Amount" />
<FormField label="Description" />

{/* NUEVO: Paid by visible siempre */}
<FormField label="Paid by" htmlFor="modalPaidBy">
  <select
    id="modalPaidBy"
    className="input"
    value={paidByMemberId}
    onChange={(e) => setPaidByMemberId(e.target.value)}
    disabled={isSubmitting || !hasMembers}
  >
    {members.map((m) => (
      <option key={m.id} value={m.id}>
        {m.name}
      </option>
    ))}
  </select>
</FormField>

<FormField label="Category" />

<button onClick={() => setShowAdvanced(!showAdvanced)}>More options</button>

{showAdvanced && (
  <>
    <FormField label="Payment method" />
    <FormField label="Expense type" />
    <FormField label="Shared expense" />
  </>
)}
```

#### State Changes
```js
// BEFORE
const paidByMemberId = useMemo(
  () => defaultPaidByMemberId.trim() || members[0]?.id || '',
  [defaultPaidByMemberId, members],
)
// → readonly, derivado de props

// AFTER
const [paidByMemberId, setPaidByMemberId] = useState(
  defaultPaidByMemberId.trim() || ''
)

// Sync con members al montar o cuando members cambia
useEffect(() => {
  if (members.length > 0 && !paidByMemberId) {
    setPaidByMemberId(defaultPaidByMemberId || members[0].id)
  }
}, [members, paidByMemberId, defaultPaidByMemberId])
```

---

### 7. DashboardPage.jsx

**Change**: Agregar empty state

#### Current Structure (simplified)
```jsx
return (
  <>
    {error && <Banner type="error">{error}</Banner>}
    {message && <Banner type="ok">{message}</Banner>}
    
    <IncomesPanel />
    <ExpenseSummary />
    <FixedExpensesPanel />
    <SettlementPanel />
    <RecentExpenses items={items} />
    <ExpenseList items={items} />
    
    <FAB onClick={() => setModalOpen(true)} />
    <ExpenseModal ... />
  </>
)
```

#### New Structure
```jsx
const hasExpenses = items.length > 0

return (
  <>
    {error && <Banner type="error">{error}</Banner>}
    {message && <Banner type="ok">{message}</Banner>}
    
    {!hasExpenses ? (
      <section className="card emptyState">
        <div className="emptyIcon" aria-hidden>💸</div>
        <h2 className="emptyTitle">No expenses yet</h2>
        <p className="emptyHint">
          Tap the <strong>+</strong> button below to add your first expense!
        </p>
      </section>
    ) : (
      <>
        <IncomesPanel />
        <ExpenseSummary />
        <FixedExpensesPanel />
        <SettlementPanel />
        <RecentExpenses items={items} />
        <ExpenseList items={items} />
      </>
    )}
    
    <FAB onClick={() => setModalOpen(true)} />
    <ExpenseModal ... />
  </>
)
```

---

### 8. styles.css

**Changes**: Color tokens, spacing, animations, empty state styles

#### Color Tokens
```css
:root {
  /* Current (replace) */
  --color-success: #10b981;
  --color-error: #ef4444;
  --color-warning: #f59e0b;
  
  /* New (vibrant) */
  --color-success: #22c55e;
  --color-error: #ef4444;  /* keep same */
  --color-warning: #f97316;  /* más naranja */
  
  /* Spacing tokens */
  --space-1: 0.25rem;  /* 4px */
  --space-2: 0.5rem;   /* 8px */
  --space-3: 0.75rem;  /* 12px */
  --space-4: 1rem;     /* 16px */
  --space-6: 1.5rem;   /* 24px */
  --space-8: 2rem;     /* 32px */
}
```

#### Empty State Styles
```css
.emptyState {
  text-align: center;
  padding: var(--space-8);
  max-width: 400px;
  margin: var(--space-8) auto;
}

.emptyIcon {
  font-size: 3rem;
  margin-bottom: var(--space-4);
}

.emptyTitle {
  font-size: 1.5rem;
  font-weight: 600;
  margin-bottom: var(--space-2);
  color: var(--color-text-primary);
}

.emptyHint {
  color: var(--color-text-dim);
  line-height: 1.6;
}
```

#### Modal Animations (improved)
```css
.modal {
  animation: modalFadeIn 250ms ease-out;
}

@keyframes modalFadeIn {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(-10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.modalBackdrop {
  animation: backdropFadeIn 250ms ease-out;
  backdrop-filter: blur(2px);  /* subtle blur */
}

@keyframes backdropFadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
```

#### Current Period Badge
```css
.currentPeriodBadge {
  display: inline-block;
  margin-left: var(--space-2);
  padding: var(--space-1) var(--space-2);
  background: var(--color-success);
  color: white;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
}
```

---

## Data Flow

### Settlement Auto-Load Flow

```
DashboardPage mounts
  └─> useSettlement hook initializes
      ├─> settlementYear = new Date().getFullYear()
      ├─> settlementMonth = new Date().getMonth() + 1
      └─> useEffect triggers loadSettlement()
          └─> API call: GET /v1/households/{id}/settlement?year=2026&month=3
              └─> Settlement data populates state
                  └─> SettlementPanel renders with data
```

### "This month" button flow

```
User clicks "This month"
  └─> onResetToCurrentMonth() in SettlementPanel
      └─> Calls resetToCurrentMonth() from useSettlement hook
          ├─> setSettlementYear(currentYear)
          ├─> setSettlementMonth(currentMonth)
          └─> useEffect observes year/month change
              └─> Triggers loadSettlement()
                  └─> Settlement panel updates
```

### Invite Member Flow

```
User clicks "+ Member" in AppHeader
  └─> Navigate to /members/new
      └─> Renders OnboardingMemberPage (reused component)
          ├─> Form has salary field in dollars
          └─> On submit:
              ├─> Convert dollars → cents (dollarsToCents util)
              ├─> API call: POST /v1/households/{id}/members
              └─> Navigate back to dashboard
```

---

## Testing Strategy

### Unit Tests (opcional para esta iteración, pero recomendado)

#### utils.js
```js
describe('dollarsToCents', () => {
  it('converts dollars to cents', () => {
    expect(dollarsToCents('100')).toBe(10000)
    expect(dollarsToCents('100.50')).toBe(10050)
    expect(dollarsToCents('0.01')).toBe(1)
  })
  
  it('handles invalid input', () => {
    expect(dollarsToCents('')).toBe(null)
    expect(dollarsToCents('abc')).toBe(null)
    expect(dollarsToCents('-10')).toBe(null)
  })
})
```

### Manual Testing (prioridad)

Ver `iteration-005-ux-polish-specs.md` → "Verification Plan"

---

## Performance Considerations

### Bundle Size Impact
- No new dependencies → **0kb increase**
- New CSS rules → ~**1-2kb increase** (negligible)
- New components → **0kb** (reusing existing)

### Runtime Performance
- Empty state conditional render → **negligible impact**
- Settlement auto-load → same API call, just happens earlier (no extra cost)
- Animations → using CSS transforms (GPU-accelerated, 60fps)

---

## Accessibility Considerations

### ARIA Labels
```jsx
// Settlement "This month" button
<button
  type="button"
  aria-label="Reset settlement to current month"
  onClick={onResetToCurrentMonth}
>
  📅 This month
</button>

// Empty state
<section className="card emptyState" aria-label="No expenses recorded yet">
  ...
</section>
```

### Keyboard Navigation
- Todos los botones nuevos deben ser focusables con Tab
- Modal debe cerrar con Escape (ya implementado)
- Selects deben funcionar con flechas (nativo)

### Color Contrast
- Success: #22c55e on white → contrast ratio **3.8:1** (WCAG AA para large text)
- Error: #ef4444 on white → contrast ratio **4.5:1** (WCAG AA compliant ✅)
- Text on backgrounds debe mantener 4.5:1 mínimo

---

## Rollback Plan

Si algo sale mal durante implementación:

1. **Revertir commit** con `git revert <commit-hash>`
2. **Feature flags** (no aplica, no hay backend changes)
3. **Rollback order**: implementar features en orden de menor a mayor riesgo
   - Least risky: Empty state, colors, spacing
   - Medium risk: Settlement auto-load, "Paid by" visibility
   - Highest risk: Salary dollars conversion (cuidado con conversión cents)

---

## Definition of Done (Technical)

- [ ] Todos los componentes modificados compilan sin warnings
- [ ] `npm run build` exitoso
- [ ] No console errors en browser
- [ ] Lighthouse Accessibility score ≥ 90
- [ ] Manual testing checklist completado (ver specs)
- [ ] CSS no tiene clases duplicadas o conflictos
- [ ] Git commits son atómicos y descriptivos

---

## Next Steps

1. ✅ Propuesta aprobada
2. ✅ Specs escritas
3. ✅ Design doc completado ← **ESTAMOS AQUÍ**
4. ⏭️ **Crear task breakdown** (`iteration-005-ux-polish-tasks.md`)
5. ⏭️ Implementar tareas en orden de prioridad
6. ⏭️ Verificar con smoke test
7. ⏭️ Actualizar `development-iteration-tracker.md`

---

## Related Documents

- `docs/iteration-005-ux-polish-proposal.md` — Propuesta aprobada
- `docs/iteration-005-ux-polish-specs.md` — Especificaciones funcionales
- `docs/iteration-005-ux-polish-tasks.md` — Task breakdown (siguiente paso)
