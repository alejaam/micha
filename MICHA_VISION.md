# Micha — Product Vision & Roadmap

## ¿Qué es Micha?

Micha es una app de finanzas colaborativas para hogares. **No es un gestor de finanzas personales** — es una herramienta para que N personas que comparten gastos tengan claridad, equidad y un cierre limpio cada periodo.

La inspiración viene de un sistema real en Excel donde dos personas con sueldos distintos contribuyen proporcionalmente a los gastos del hogar, registran sus propios gastos, y al final del mes calculan quién pagó de más y liquidan la diferencia. Micha digitaliza y generaliza esa lógica para cualquier configuración de hogar.

---

## Conceptos Clave (Glosario — leer antes de codear)

| Concepto | Definición |
|---|---|
| **Household** | Unidad de gestión compartida. Es la "cuenta" o "casa". Un usuario puede pertenecer a múltiples Households (ej. "Casa Roto", "Depa CDMX"). Tiene un owner con permisos especiales. |
| **Member** | Persona dentro de un Household. Tiene un % de contribución asignado manualmente. |
| **% de Contribución** | Proporción del gasto compartido que le corresponde a cada miembro. Se define manualmente, se puede actualizar en cualquier momento, pero **solo aplica hacia adelante** — nunca modifica periodos pasados. |
| **Periodo** | Ciclo financiero del Household. Duración configurable (mensual, quincenal, personalizado). Tiene tres estados posibles: `open`, `review`, `closed`. Solo puede existir un periodo `open` por Household a la vez. |
| **Ingreso** | El sueldo o entrada de dinero de cada miembro en un periodo. Se usa para calcular el % de contribución sugerido, pero el % final lo define el usuario manualmente. |
| **Gasto Fijo** | Gasto recurrente predefinido (renta, internet, streaming). Se replica automáticamente al abrir cada nuevo periodo. Tiene asignación proporcional entre miembros según su %. |
| **Gasto Variable** | Gasto registrado por un miembro durante el periodo activo. No es predecible mes a mes. Lo registra cada miembro para sí mismo. |
| **MSI / Cuota** | Gasto a meses sin intereses. **No existe como gasto único en el mes de compra** — existe como cuota mensual recurrente. Se modela como una entidad separada (`Installment`) con: monto total, número de cuotas, cuota actual, fecha de inicio. Se replica automáticamente en cada nuevo periodo hasta completarse. |
| **Balance / Ajuste** | Cálculo en tiempo real de quién pagó de más o de menos respecto a su % de contribución. Se recalcula con cada gasto nuevo. Se **liquida y resetea en cada cierre** — no se acumula entre periodos. |
| **Cierre de Periodo** | Acto formal y consensuado. Flujo: cualquier miembro inicia el cierre → periodo pasa a `review` → cada miembro aprueba ✅ o objeta ❌ con comentario → si hay consenso se cierra, si hay objeciones el owner del Household tiene la última palabra → al cerrar se genera automáticamente el siguiente periodo. |
| **Panel** | Sección visual del dashboard que agrupa un tipo de información. La UI principal es un dashboard de paneles densos — no una lista de navegación. El usuario debe ver el estado del periodo de un vistazo sin navegar entre pantallas. |

---

## Reglas de Negocio Críticas

1. El **% de contribución nunca es retroactivo**. Si cambias el % hoy, los periodos anteriores conservan el % que tenían.
2. Una **cuota MSI vive en el mes que se cobra**, no en el mes en que se hizo la compra.
3. El **balance se recalcula en tiempo real** con cada gasto registrado.
4. El **cierre requiere acción humana** — no ocurre automáticamente al terminar el periodo.
5. Durante el estado `review` **no se pueden agregar gastos nuevos**.
6. El **owner del Household** puede forzar el cierre aunque haya objeciones.
7. Al cerrar un periodo, los **gastos fijos y cuotas MSI activas se replican** automáticamente en el nuevo periodo.
8. Cada miembro **registra sus propios gastos variables**.
9. Micha **no gestiona finanzas personales** — solo lo que se comparte en el Household.

---

## Arquitectura de Datos (Entidades)

Estas son las entidades que debe tener el modelo de datos. No agregues ni quites sin consultar el documento de visión.

- **User** — cuenta de usuario
- **Household** — nombre, owner (User), ciclo (enum: monthly | biweekly | custom), fecha de creación
- **HouseholdMember** — relación User ↔ Household, porcentaje de contribución, fecha desde la que aplica ese %
- **Period** — Household, fecha inicio, fecha fin, estado (open | review | closed)
- **Income** — miembro, periodo, monto
- **Category** — nombre, tipo (predefined | custom), Household (null si es predefinida)
- **Expense** — miembro, periodo, monto, categoría, descripción, tipo (fixed | variable | installment), fecha
- **Installment** — Expense origen, monto total, total de cuotas, número de cuota actual, fecha de inicio
- **PeriodApproval** — miembro, periodo, estado (approved | objected), comentario, timestamp
- **Balance** — derivado/calculado: miembro, periodo, total pagado, total que debería haber pagado, diferencia

---

## Categorías Predefinidas

El sistema incluye estas categorías por defecto. El usuario puede crear las suyas propias por Household:

- Vivienda (Renta, Agua, Luz, Gas)
- Transporte (Gasolina, Uber, Casetas)
- Alimentación (Mercado, Restaurantes, Delivery)
- Streaming & Servicios (Netflix, Spotify, Internet, Suscripciones)
- Salud
- Entretenimiento
- Ropa & Personal
- Otros

---

## UI — Principios

1. **Densidad sobre navegación** — La información importante debe ser visible sin hacer tap adicional
2. **Tiempo real** — El balance nunca debe estar desactualizado
3. **Paneles segmentados** — El dashboard se divide en bloques visuales claros, cada uno con su tema
4. **Consenso explícito** — El cierre requiere acción activa de los miembros

### Paneles del Dashboard (orden)

1. **Ingresos** — ingreso declarado por miembro + % de contribución
2. **Balance Actual** — quién debe cuánto, en tiempo real
3. **Gastos Fijos** — recurrentes del periodo con asignación proporcional
4. **Tarjetas / MSI** — cuotas activas con progreso (ej. 11 de 24)
5. **Gastos Variables** — todos los gastos del periodo, agrupados por miembro o categoría
6. **Cierre de Periodo** — estado de aprobación de cada miembro

---

## Fases de Desarrollo

### FASE 0 — Modelo de Datos ✅ COMPLETADA
Crear el esquema de base de datos con todas las entidades del glosario. Sin UI. Validar que las reglas de negocio críticas estén representadas en el modelo.

**Criterios:**
- [x] Todas las entidades listadas existen con sus campos
- [x] HouseholdMember tiene fecha de vigencia del %
- [x] Installment es una entidad separada de Expense
- [x] Period tiene los tres estados correctos
- [x] Balance es calculado/derivado, no un campo editable

**Estado de implementación:**
- ✅ Entidades de dominio creadas en `backend/internal/domain/`:
  - `household/` — Household con owner, cycleType (monthly|biweekly|custom)
  - `member/` — Member con contributionPct y validFrom (no retroactivo)
  - `period/` — Period con status (open|review|closed)
  - `income/` — Income vinculado a member + period
  - `category/` — Category con categoryType (predefined|custom)
  - `expense/` — Expense actualizado con memberID, periodID, categoryID
  - `installment/` — Installment separado de Expense con tracking de cuotas
  - `period_approval/` — PeriodApproval con status (approved|objected)
- ✅ Reglas de negocio documentadas en `backend/internal/domain/rules.go`
- ✅ Errores de dominio centralizados en `backend/internal/domain/shared/errors.go`
- ⚠️ **Pendiente:** Actualizar capas superiores (application, ports, adapters) para alinearse con el nuevo modelo de dominio

### FASE 1 — Dashboard de Paneles
Implementar la pantalla principal con los 6 paneles. El balance debe actualizarse en tiempo real al registrar un gasto.

### FASE 2 — Registro de Gastos
Flujo para registrar gastos variables, fijos y MSI. Al guardar, el balance se recalcula.

### FASE 3 — Cierre de Periodo
Flujo completo: iniciar cierre → revisión → aprobaciones → forzar si hay objeciones → cerrar → generar nuevo periodo.

### FASE 4 — Gestión de Household
Crear household, invitar miembros, definir %, transferir ownership, archivar.

---

## Lo que Micha NO es (límites del producto por ahora)

- ❌ No es un gestor de finanzas personales
- ❌ No da consejos ni alertas de ahorro
- ❌ No conecta con bancos ni APIs financieras
- ❌ No proyecta ni predice gastos futuros
- ❌ No es proactiva — solo reactiva (registro y consulta)

---

## Estado del Proyecto

### Arquitectura Implementada
- **Backend:** Go 1.23+ con DDD + Clean Architecture + Hexagonal
- **Estructura de capas:**
  - `domain/` — Entidades puras sin dependencias externas ✅
  - `application/` — Casos de uso (requiere actualización)
  - `ports/` — Contratos inbound/outbound (requiere actualización)
  - `adapters/` — HTTP + Postgres (requiere actualización)
  - `infrastructure/` — Config, DB, migraciones

### Próximos Pasos
1. Actualizar application layer para usar nuevas entidades de dominio
2. Actualizar ports (inbound/outbound) para reflejar nuevos contratos
3. Actualizar adapters (HTTP handlers + Postgres repositories)
4. Crear/actualizar migraciones de PostgreSQL
5. Actualizar tests para nuevas entidades

---

*Documento vivo. Consultar antes de cada tarea. Actualizar al inicio de cada fase.*
*Versión: 0.2 — Marzo 2026 — FASE 0 completada*
