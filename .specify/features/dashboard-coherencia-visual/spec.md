     1|# Revisión de coherencia y visualización de dashboards
     2|
     3|## Contexto
     4|Queremos revisar y mejorar la **coherencia** y la **visualización** de todos los dashboards / vistas tipo dashboard de la app (front-end), asegurando consistencia en layout, jerarquía visual, componentes, accesibilidad, estados vacío/carga/error y patrones de interacción.
     5|
     6|**Alcance inferido (según el repo actual):**
     7|- `frontend/src/pages/DashboardPage.jsx`
     8|- `frontend/src/pages/ExpensesPage.jsx` (incluye resumen + history; layout tipo dashboard)
     9|- `frontend/src/pages/BalancesPage.jsx`
    10|- `frontend/src/pages/InstallmentsPage.jsx`
    11|- `frontend/src/pages/RulesPage.jsx` (paneles tipo dashboard)
    12|- `frontend/src/pages/FixedExpensesPage.jsx` y `frontend/src/pages/OnboardingFixedExpensesPage.jsx` (paneles/tabla/móvil)
    13|
    14|## User Scenarios & Testing (S1..)
    15|
    16|### S1 · Navegación consistente entre dashboards
    17|- **Dado** que navego Dashboard → Gastos → Balances → Reglas → MSI
    18|- **Cuando** cambio de página
    19|- **Entonces** el layout (grid/columnas), espaciados, tipografía, estilo de cards, headers y CTAs se perciben consistentes.
    20|
    21|### S2 · Jerarquía de información clara
    22|- **Dado** un dashboard
    23|- **Cuando** lo abro
    24|- **Entonces** veo primero estado del periodo/resumen global, luego detalle (listas/paneles), y por último acciones.
    25|
    26|### S3 · Estados vacíos/loading/error consistentes
    27|- **Dado** que no hay datos / hay carga / hay error
    28|- **Cuando** abro cualquier dashboard
    29|- **Entonces** el estilo de EmptyState/Banner/loading no cambia de manera inesperada entre páginas.
    30|
    31|### S4 · Usabilidad móvil
    32|- **Dado** un viewport pequeño
    33|- **Cuando** abro dashboards
    34|- **Entonces** no hay overflow lateral, el scroll es predecible, y los CTAs son alcanzables.
    35|
    36|### S5 · Accesibilidad consistente
    37|- **Dado** un lector de pantalla o navegación por teclado
    38|- **Cuando** recorro dashboards
    39|- **Entonces** los landmarks/aria-labels/orden DOM y estados de foco son coherentes.
    40|
    41|## Requirements (R1..)
    42|
    43|### R1 · Inventario y definición de “dashboard”
    44|- Debe existir un inventario explícito de páginas/paneles que cuentan como “dashboard”.
    45|
    46|### R2 · Checklist de coherencia visual
    47|- Debe existir una checklist con al menos:
    48|  - Layout: `pageGrid` vs `dashboardCol`
    49|  - Uso consistente de `card`, `listHeader`, `listTitle`, `listCount`, `sectionTitle`
    50|  - Banners: `error/ok`, floating vs inline
    51|  - Empty states
    52|  - Densidad (padding/margins/gaps) y tipografía
    53|
    54|### R3 · Reglas de jerarquía y orden
    55|- Debe definirse y aplicarse una regla de orden: **estado del periodo → resumen → detalle → acciones**.
    56|- Debe respetarse el orden DOM cuando sea relevante para a11y (hay tests existentes).
    57|
    58|### R4 · Consistencia de interacciones
    59|- FAB / Quick Add / Modals / BottomSheet deben comportarse de forma consistente entre dashboards.
    60|
    61|### R5 · Accesibilidad mínima
    62|- Secciones principales con `aria-label` y/o headings consistentes.
    63|- Mantener o ajustar tests existentes (hay un test de orden DOM del dashboard).
    64|
    65|### R6 · Verificación objetiva (Definition of Done)
    66|- Debe haber verificación objetiva: tests que pasen, checklist marcada, y evidencia por dashboard (p. ej. notas de revisión o snapshots).
    67|
    68|## Success Criteria (SC1..)
    69|- SC1: Existe documento de inventario + checklist y se aplica a las páginas objetivo.
    70|- SC2: No hay regresiones de layout (móvil/desktop) en las páginas revisadas.
    71|- SC3: Estados de error/empty/loading son consistentes.
    72|- SC4: Tests relevantes pasan y/o se agregan donde falten.
    73|
    74|## Out of Scope
    75|- Rediseño total del sistema de diseño.
    76|- Cambios de negocio en cálculos financieros.
    77|
    78|## Open Questions
(Resueltas)
- Repo objetivo: **/opt/data/repos/micha** (confirmado)
- Objetivo: **implementar cambios UI** (no solo auditoría)
- Alcance: por ahora solo mejorar pantallas para que **no se sientan onboarding** y asegurar navegación (no depender de "Saltar por ahora")

Nuevas preguntas:
1) ¿Qué navegación esperas exactamente? (p. ej. botón "Volver" a Settings/Rules o a Dashboard, y breadcrumb/back del navegador)
2) ¿Qué pantallas específicas hoy se sienten "onboarding" para ti? (Dashboard vacío, Cards, Fixed Expenses, otras)
3) ¿Quieres mantener el flujo onboarding, pero con UI distinta, o eliminar onboarding routes cuando ya hay household configurado?
