# Iteración 005 — UX Polish — Task Breakdown

**Propuesta**: `iteration-005-ux-polish-proposal.md`  
**Specs**: `iteration-005-ux-polish-specs.md`  
**Design**: `iteration-005-ux-polish-design.md`  
**Estado**: Task Breakdown  
**Fecha**: 2026-03-24

---

## Task Organization

Las tareas están organizadas por:
1. **Priority** (P0 = Must Have, P1 = Should Have, P2 = Nice to Have)
2. **Phase** (Setup, Implementation, Polish, Verification)
3. **Estimated time** (S = Small <30min, M = Medium 30-60min, L = Large 1-2hr)

---

## Phase 0: Setup & Preparation

### T0.1 — Create feature branch
- **Priority**: P0
- **Time**: S (5 min)
- **Description**: Create branch `feature/iteration-005-ux-polish` from `main`
- **Commands**:
  ```bash
  git checkout main
  git pull
  git checkout -b feature/iteration-005-ux-polish
  ```

### T0.2 — Verify frontend builds
- **Priority**: P0
- **Time**: S (5 min)
- **Description**: Ensure clean build before starting
- **Commands**:
  ```bash
  cd frontend
  npm install
  npm run build
  npm run dev  # verify localhost:5173 works
  ```

---

## Phase 1: Core UX Fixes (Must Have)

### T1.1 — Add dollarsToCents utility
- **Priority**: P0
- **Time**: S (15 min)
- **Files**: `frontend/src/utils.js`
- **Description**: 
  - Add `dollarsToCents(dollars)` function
  - Handle decimals, invalid input, negative numbers
  - Return `null` for invalid input
- **Test manually**:
  ```js
  dollarsToCents('100')      // → 10000
  dollarsToCents('100.50')   // → 10050
  dollarsToCents('invalid')  // → null
  ```

### T1.2 — Update OnboardingMemberPage salary field
- **Priority**: P0
- **Time**: M (30 min)
- **Files**: `frontend/src/pages/OnboardingMemberPage.jsx`
- **Changes**:
  1. Change state from `salary` to `salaryDollars`
  2. Update label: "Monthly salary (cents, optional)" → "Monthly salary (optional)"
  3. Add input prefix `$` (use `.inputWrap` + `.inputPrefix`)
  4. Add `step="0.01"` to input
  5. Update placeholder: "0" → "e.g., 5000"
  6. Convert to cents before API call: `dollarsToCents(salaryDollars)`
- **Verify**:
  - Input "5000" → backend receives `500000`
  - Input "5000.50" → backend receives `500050`
  - Input "" (empty) → backend receives `0`

### T1.3 — Settlement auto-load current month
- **Priority**: P0
- **Time**: M (30 min)
- **Files**: `frontend/src/hooks/useSettlement.js`
- **Changes**:
  1. Initialize `settlementYear` with `new Date().getFullYear()`
  2. Initialize `settlementMonth` with `new Date().getMonth() + 1`
  3. Add `useEffect` to auto-call `loadSettlement()` on mount
  4. Add `resetToCurrentMonth()` function
  5. Return `resetToCurrentMonth` in hook return value
- **Verify**:
  - Dashboard loads → settlement shows current March 2026
  - Manual change to Jan 2025 → data updates
  - No infinite loop (check useEffect dependencies)

### T1.4 — Add "This month" button to SettlementPanel
- **Priority**: P0
- **Time**: S (20 min)
- **Files**: `frontend/src/components/SettlementPanel.jsx`
- **Changes**:
  1. Add prop `onResetToCurrentMonth`
  2. Add button after "Refresh" button
  3. Add `isCurrentMonth` check
  4. Add `<span className="currentPeriodBadge">(current)</span>` in header if current
- **Verify**:
  - Button visible and clickable
  - Click → resets to current month
  - Badge shows "(current)" when on current month

### T1.5 — Wire "This month" in DashboardPage
- **Priority**: P0
- **Time**: S (10 min)
- **Files**: `frontend/src/pages/DashboardPage.jsx`
- **Changes**:
  1. Destructure `resetToCurrentMonth` from `useSettlement`
  2. Pass as `onResetToCurrentMonth` prop to `SettlementPanel`
- **Verify**:
  - End-to-end: Dashboard → change month → click "This month" → resets

### T1.6 — Move "Paid by" field out of "More options"
- **Priority**: P0
- **Time**: M (40 min)
- **Files**: `frontend/src/components/ExpenseModal.jsx`
- **Changes**:
  1. Change `paidByMemberId` from `useMemo` to `useState`
  2. Add `useEffect` to sync with `members` and `defaultPaidByMemberId`
  3. Move `<FormField label="Paid by">` outside `showAdvanced` block
  4. Convert from `<p>Paid by: {name}</p>` to `<select>` dropdown
  5. Position after "Description", before "Category"
- **Verify**:
  - "Paid by" visible immediately
  - Dropdown shows all members
  - Auto-selects current member
  - Changing selection works

### T1.7 — Add empty state to DashboardPage
- **Priority**: P0
- **Time**: M (30 min)
- **Files**: 
  - `frontend/src/pages/DashboardPage.jsx`
  - `frontend/src/styles.css`
- **Changes**:
  1. Add `hasExpenses = items.length > 0` check
  2. Conditional render: if `!hasExpenses`, show empty state
  3. Empty state structure: icon 💸, title, hint text
  4. Add CSS for `.emptyState`, `.emptyIcon`, `.emptyTitle`, `.emptyHint`
- **Verify**:
  - New user → sees empty state
  - Add first expense → empty state disappears
  - Text is friendly and actionable

---

## Phase 2: Secondary Improvements (Should Have)

### T2.1 — Add "+ Member" button to AppHeader
- **Priority**: P1
- **Time**: M (30 min)
- **Files**: `frontend/src/components/AppHeader.jsx`
- **Changes**:
  1. Import `Link` from `react-router-dom`
  2. Add button/link after HouseholdSelector
  3. Style with `.btn .btnSm .btnGhost`
  4. Label: "+ Member" or "Invite Member"
- **Verify**:
  - Button visible in header
  - Click → navigates to `/members/new`

### T2.2 — Add route /members/new
- **Priority**: P1
- **Time**: S (10 min)
- **Files**: `frontend/src/router.jsx`
- **Changes**:
  1. Add route: `<Route path="/members/new" element={<OnboardingMemberPage />} />`
- **Verify**:
  - Navigate to `/members/new` → renders OnboardingMemberPage
  - Form works identically to onboarding flow
  - After submit → redirects to dashboard

### T2.3 — Update color tokens (vibrant palette)
- **Priority**: P1
- **Time**: S (15 min)
- **Files**: `frontend/src/styles.css`
- **Changes**:
  1. Update `--color-success` to `#22c55e`
  2. Update `--color-warning` to `#f97316`
  3. Update `--color-error` (keep `#ef4444`)
  4. Add spacing tokens (`--space-1` through `--space-8`)
- **Verify**:
  - Success banners use new green
  - Warning states use new orange
  - Spacing tokens defined (check in DevTools)

### T2.4 — Add spacing tokens to components
- **Priority**: P1
- **Time**: M (45 min)
- **Files**: `frontend/src/styles.css` (various classes)
- **Description**: 
  - Replace hardcoded px values with `var(--space-X)` tokens
  - Focus on: `.card`, `.formStack`, `.modalActions`, `.banner`
- **Verify**:
  - Visual regression check (components look the same)
  - DevTools shows `var(--space-4)` instead of `16px`

### T2.5 — Improve modal animations
- **Priority**: P1
- **Time**: S (20 min)
- **Files**: `frontend/src/styles.css`
- **Changes**:
  1. Update `.modal` animation: add scale + translateY
  2. Update `.modalBackdrop`: add `backdrop-filter: blur(2px)`
  3. Timing: 250ms ease-out
- **Verify**:
  - Modal fades in smoothly
  - Backdrop has subtle blur
  - No jank (60fps)

### T2.6 — Add currentPeriodBadge CSS
- **Priority**: P1
- **Time**: S (10 min)
- **Files**: `frontend/src/styles.css`
- **Changes**:
  1. Add `.currentPeriodBadge` class
  2. Style: green bg, white text, small, rounded
- **Verify**:
  - Badge renders correctly in SettlementPanel header
  - Only shows when on current month

---

## Phase 3: Polish & Nice-to-Haves (Optional)

### T3.1 — Add category icons
- **Priority**: P2
- **Time**: M (30 min)
- **Files**: 
  - `frontend/src/components/ExpenseModal.jsx`
  - `frontend/src/components/RecentExpenses.jsx` (or ExpenseItem)
- **Changes**:
  1. Create icon mapping: `{ rent: '🏠', auto: '🚗', food: '🍔', ... }`
  2. Render icon before category label in dropdown
  3. Render icon in expense cards
- **Verify**:
  - Icons show in category dropdown
  - Icons show in expense list

### T3.2 — Add tooltips to key fields
- **Priority**: P2
- **Time**: M (45 min)
- **Files**: 
  - `frontend/src/ui/Tooltip.jsx` (new component, optional)
  - OR use HTML `title` attribute
- **Changes**:
  1. Add tooltip to "Shared expense" checkbox
  2. Add tooltip to "Settlement mode" in SettlementPanel
  3. Add tooltip to "Expense type"
- **Verify**:
  - Hover → tooltip appears
  - Tooltip text is helpful
  - Accessible (ARIA labels)

### T3.3 — Add members list section
- **Priority**: P2
- **Time**: L (1-2 hr)
- **Files**: 
  - `frontend/src/pages/DashboardPage.jsx`
  - OR `frontend/src/pages/MembersPage.jsx` (new)
  - `frontend/src/components/MembersList.jsx` (new)
- **Changes**:
  1. Create `<MembersList>` component
  2. Display members: name, email, role, salary (optional)
  3. Add to dashboard OR create separate `/members` route
  4. Empty state: "You're the only member. Invite others!"
- **Verify**:
  - Members list renders correctly
  - Shows all household members
  - Empty state shows when only one member

---

## Phase 4: Verification & Documentation

### T4.1 — Run smoke test
- **Priority**: P0
- **Time**: M (30 min)
- **Description**: Execute full smoke test flow (see specs)
- **Steps**:
  1. Register new user
  2. Create household
  3. Verify empty state
  4. Add first member (check salary field)
  5. Add first expense (check "Paid by" visible)
  6. Verify settlement auto-loads current month
  7. Click "This month" button
  8. Add second member via "+ Member" button
  9. Add second expense
  10. Verify colors, animations, spacing
- **Checklist**: Use manual testing checklist from specs

### T4.2 — Build for production
- **Priority**: P0
- **Time**: S (5 min)
- **Commands**:
  ```bash
  cd frontend
  npm run build
  ```
- **Verify**:
  - Build succeeds with no errors
  - No console warnings
  - `dist/` folder generated

### T4.3 — Mobile responsive check
- **Priority**: P0
- **Time**: S (15 min)
- **Description**: Test in mobile viewport (375px width)
- **Verify**:
  - All modals fit on screen
  - Buttons are tappable (min 44x44px)
  - Text is readable
  - No horizontal scroll

### T4.4 — Update development-iteration-tracker.md
- **Priority**: P0
- **Time**: M (30 min)
- **Files**: `docs/development-iteration-tracker.md`
- **Changes**:
  1. Add "Iteración 005 — UX Polish" section
  2. Fill template: Fecha, Fase, MEM, SEQ, THINK
  3. List all files changed
  4. Document decisions made
  5. List risks/deuda remaining
- **Verify**:
  - Tracker updated
  - Next person can understand what was done

### T4.5 — Create PR or merge to main
- **Priority**: P0
- **Time**: S (10 min)
- **Commands**:
  ```bash
  git add .
  git commit -m "feat: iteration 005 - UX polish

  - Salary field in dollars (not cents)
  - Settlement auto-loads current month
  - Add 'This month' button
  - Move 'Paid by' field to main form
  - Add empty state for no expenses
  - Add '+ Member' button in header
  - Improve color palette and animations

  Closes #XXX"
  
  git push origin feature/iteration-005-ux-polish
  # Then create PR on GitHub
  ```

---

## Task Summary

| Phase | P0 (Must) | P1 (Should) | P2 (Nice) | Total |
|-------|-----------|-------------|-----------|-------|
| Phase 0 | 2 | 0 | 0 | 2 |
| Phase 1 | 7 | 0 | 0 | 7 |
| Phase 2 | 0 | 6 | 0 | 6 |
| Phase 3 | 0 | 0 | 3 | 3 |
| Phase 4 | 5 | 0 | 0 | 5 |
| **Total** | **14** | **6** | **3** | **23** |

### Time Estimates

- **P0 tasks**: ~5 hours
- **P1 tasks**: ~3 hours
- **P2 tasks**: ~3 hours
- **Total (all tasks)**: ~11 hours (~1.5 days de trabajo concentrado)

### Recommended Order (by priority)

**Day 1 (Core - P0)**:
1. T0.1, T0.2 (setup)
2. T1.1, T1.2 (salary in dollars)
3. T1.3, T1.4, T1.5 (settlement auto-load)
4. T1.6 (Paid by field)
5. T1.7 (empty state)

**Day 2 (Polish - P1)**:
6. T2.1, T2.2 (Invite Member button)
7. T2.3, T2.4 (color tokens + spacing)
8. T2.5, T2.6 (animations + badge)

**Day 3 (Optional + Verification)**:
9. T3.1, T3.2, T3.3 (nice-to-haves if time permits)
10. T4.1, T4.2, T4.3 (smoke test + verification)
11. T4.4, T4.5 (documentation + merge)

---

## Risks & Blockers

| Task | Risk | Mitigation |
|------|------|------------|
| T1.2 | Romper onboarding flow | Test thoroughly before moving to next task |
| T1.3 | useEffect infinite loop | Careful with dependencies, test immediately |
| T1.6 | Breaking expense creation | Smoke test después de este cambio |
| T4.1 | Descubrir bugs tarde | Hacer smoke test incremental durante dev |

---

## Definition of Done Checklist

- [ ] All P0 tasks completed
- [ ] At least 80% of P1 tasks completed
- [ ] Smoke test passed (T4.1)
- [ ] Production build succeeds (T4.2)
- [ ] Mobile responsive (T4.3)
- [ ] Documentation updated (T4.4)
- [ ] No console errors in browser
- [ ] Git commits are clean and descriptive
- [ ] Ready to merge to main

---

## Next Steps

**¿Estás listo para empezar la implementación?**

Si la respuesta es **SÍ**:
1. Ejecuta T0.1 (crear branch)
2. Ejecuta T0.2 (verificar build)
3. Comienza con T1.1 (dollarsToCents utility)
4. Trabaja en orden, marcando tasks como completadas

Si la respuesta es **NO** (quieres revisar primero):
- Revisa esta propuesta completa
- Haz preguntas o ajustes
- Aprueba formalmente antes de iniciar

---

## Related Documents

- `docs/iteration-005-ux-polish-proposal.md` — Propuesta aprobada
- `docs/iteration-005-ux-polish-specs.md` — Especificaciones funcionales
- `docs/iteration-005-ux-polish-design.md` — Diseño técnico
- `docs/development-iteration-tracker.md` — Para actualizar al finalizar
