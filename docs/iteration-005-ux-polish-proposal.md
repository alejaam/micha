# Iteración 005 — UX Polish

**Estado**: Propuesta  
**Fecha**: 2026-03-24  
**Fase**: Intermedia (entre Phase 3 y Phase 4)  
**Duración estimada**: ~1 semana (5-7 días)  
**Alcance**: Frontend únicamente

---

## 1. Intent (Qué y Por qué)

### Problema
La app micha tiene un backend sólido (A+) y un frontend funcional, pero presenta **gaps críticos de UX** que hacen que la experiencia se sienta inacabada y confusa para el usuario:

1. **Onboarding abrumador**: Campo "Monthly salary (cents)" confunde (debería ser dólares)
2. **Settlement oculto**: Requiere selección manual de año/mes, debería auto-mostrar mes actual
3. **Sin gestión de miembros**: No hay botón "Invite Member" después del onboarding
4. **Campos importantes ocultos**: "Paid by", "Payment Method", "Expense Type" están bajo "More options"
5. **Falta feedback visual**: Sin estados vacíos, sin tooltips, sin confirmaciones visuales
6. **Dashboard denso**: 6 paneles sin explicación para nuevos usuarios

### Objetivo
Pulir la experiencia de usuario **antes de agregar más features** (Phase 4). Crear una base UX sólida que haga la app delightful y prepare el terreno para features futuras.

### Impacto esperado
- Reducir fricción en onboarding (50% menos confusión)
- Aumentar engagement (usuarios completan primera acción más rápido)
- Mejorar retención (usuarios entienden el valor inmediatamente)

---

## 2. Scope

### ✅ In Scope (Frontend only)

#### A. Onboarding improvements
- Cambiar "Monthly salary (cents)" → "Monthly salary" con campo en dólares
- Agregar tooltips explicativos en campos clave
- Mejorar copy de instrucciones

#### B. Settlement auto-period
- Auto-seleccionar mes/año actual al cargar dashboard
- Agregar botón "This month" para resetear rápidamente
- Mostrar período actual en el header del panel

#### C. Member management UI
- Agregar botón "Invite Member" en dashboard o app header
- Crear ruta/modal para invitar miembros post-onboarding
- Mostrar lista de miembros actuales

#### D. Expense form visibility
- Hacer "Paid by" siempre visible (no en "More options")
- Considerar hacer "Payment Method" visible por defecto
- Agregar hint: "Paid by you" cuando auto-selecciona current member

#### E. Empty states & feedback
- Agregar empty state en dashboard: "No expenses yet. Tap + to add your first one!"
- Agregar confirmación visual al crear/editar/eliminar gasto (toast mejorado)
- Agregar tooltips en paneles principales (icon con "?")

#### F. Visual polish
- Mejorar spacing y jerarquía visual
- Agregar iconos a categorías
- Mejorar colores de estados (success, error, warning)
- Pulir animaciones de modales y transiciones

### ❌ Out of Scope

- Backend changes (el backend está sólido, no requiere cambios)
- Nuevas features (gastos recurrentes, MSI, etc. → Phase 4)
- Concepto de "períodos" explícitos (open/review/closed → decisión posterior)
- Autenticación OAuth
- Reportes avanzados

---

## 3. Approach (Cómo)

### Estrategia de implementación
1. **Orden**: Priorizar cambios por impacto (onboarding primero, polish visual último)
2. **Incremental**: Cada mejora se prueba independientemente
3. **No breaking changes**: Mantener compatibilidad con backend actual
4. **Quick wins**: Cambios pequeños con alto impacto

### Tecnologías
- React 18 + Vite (stack actual)
- CSS vanilla (sin librerías nuevas)
- API existente (sin cambios)

### Plan de verificación
- [ ] Build exitoso (`npm run build`)
- [ ] Smoke test manual en localhost
- [ ] Verificar flujo completo: register → onboarding → add expense → settlement
- [ ] Probar en mobile viewport (responsive)

---

## 4. Success Criteria

### Criterios de aceptación

#### Must Have (obligatorios)
- [ ] Onboarding muestra "Monthly salary" en dólares (no cents)
- [ ] Settlement auto-selecciona mes/año actual al cargar
- [ ] Existe botón "Invite Member" accesible desde dashboard
- [ ] Campo "Paid by" visible sin expandir "More options"
- [ ] Dashboard muestra empty state cuando no hay gastos
- [ ] Build de frontend exitoso sin errores

#### Should Have (deseables)
- [ ] Tooltips explicativos en campos clave
- [ ] Botón "This month" en settlement para reset rápido
- [ ] Lista de miembros visible en alguna vista
- [ ] Iconos en categorías de gastos
- [ ] Confirmaciones visuales mejoradas (toasts)

#### Nice to Have (opcionales, si da tiempo)
- [ ] Animaciones pulidas en modales
- [ ] Tema de colores más vibrante
- [ ] Hints contextuales en paneles vacíos

---

## 5. Risks & Mitigations

| Riesgo | Probabilidad | Impacto | Mitigación |
|--------|--------------|---------|------------|
| Romper flujo existente | Baja | Alto | Smoke test completo después de cada cambio |
| Over-engineering UI | Media | Medio | Mantener cambios simples, sin librerías nuevas |
| Desvío de tiempo (>1 semana) | Media | Bajo | Priorizar "Must Have", postponer "Nice to Have" |
| Incompatibilidad con backend | Muy baja | Alto | No cambiar contratos API, solo UI |

---

## 6. Dependencies & Blockers

### Dependencies
- Ninguna (todos los endpoints necesarios ya existen)

### Blockers
- Ninguno identificado

---

## 7. Next Steps

Si esta propuesta es aprobada:

1. **Crear specs detalladas** → `iteration-005-ux-polish-specs.md`
   - User stories por cada mejora UX
   - Wireframes/mockups si es necesario
   - Acceptance criteria granulares

2. **Crear design doc** → `iteration-005-ux-polish-design.md`
   - Componentes a modificar
   - Cambios de estado
   - Decisiones técnicas (routing, state management, etc.)

3. **Break down tasks** → `iteration-005-ux-polish-tasks.md`
   - Lista de tareas agrupadas por prioridad
   - Orden de ejecución
   - Estimaciones de tiempo

4. **Implement** → Ejecutar tareas en orden
5. **Verify** → Smoke test + checklist
6. **Document** → Actualizar `development-iteration-tracker.md`

---

## 8. Approvals

- [ ] Product Owner (Ale) aprueba scope
- [ ] Engineering (Ale) confirma viabilidad técnica
- [ ] Ready para crear specs

---

## Related Documents

- `docs/product-roadmap.md` — Roadmap general del producto
- `docs/development-iteration-tracker.md` — Historial de iteraciones
- `MICHA_VISION.md` — Visión a largo plazo (períodos, MSI, etc.)
- `docs/architecture-checklist.md` — Verificación de arquitectura (no aplica para esta iteración, solo frontend)
