# Micha — Glosario del Dominio

## Términos

### Household
Unidad de gestión financiera compartida. Es el contexto raíz del sistema. Un usuario puede pertenecer a múltiples Households. Tiene un owner con permisos especiales y una configuración de liquidación (`SettlementMode`).

### Member
Persona dentro de un Household. Puede estar vinculada a un `User` (cuando acepta la invitación) o ser un registro pendiente. Cada Member tiene un salario (`MonthlySalaryCents`) que se usa para calcular su contribución esperada en modo proporcional.

### Contribución Esperada (derivado)
Porción del gasto compartido que le corresponde a cada Member. **No se almacena** — se calcula en tiempo real según:
- `SettlementMode.equal` → división equitativa (1/N)
- `SettlementMode.proportional` → proporcional al `MonthlySalaryCents` de cada Member
- `SplitConfig` → override manual que reemplaza el cálculo automático

Los cambios en la contribución **nunca son retroactivos**: solo afectan al periodo open actual y futuros.

### SettlementMode
Modo de liquidación del Household. Determina cómo se divide el gasto compartido entre los miembros:
- `equal` — todos pagan lo mismo
- `proportional` — proporcional al salario declarado

### SplitConfig
Configuración opcional de splits personalizados. Cuando existe, reemplaza al `SettlementMode` para el cálculo de contribuciones. Define porcentajes explícitos (sumando 100%) por Member. No retroactivo.

### Period
Ciclo financiero de un Household. Duración configurable (`monthly` | `biweekly`). Tres estados: `open`, `review`, `closed`. El rango de fechas se define por el `closingDay` del Household:
- Inicia: día después del closingDay (ej: closingDay=15 → arranca el 16)
- Termina: el próximo closingDay (ej: 15 del mes siguiente)

Solo puede existir un Period `open` por Household a la vez. Al cerrar un Period se genera automáticamente el siguiente.

### RecurringExpense (template)
Plantilla que genera automáticamente un `ExpenseTypeFixed` en el Period activo cuando toca según su schedule. El gasto generado **no tiene pagador asignado** — se divide directamente entre los miembros según su contribución. El `RecurringExpense` puede actualizarse si el monto cambia (ej: la luz sube), pero siempre se cuenta.

Campos: `HouseholdID`, `amountCents`, `description`, `categoryID`, `RecurrencePattern`, `StartDate`, `EndDate`, `NextGenerationDate`, `IsActive`.

### Expense
Registro de un gasto compartido dentro de un Period. No existe el concepto de gasto personal en Micha. Referencia a una `Card` por ID si se pagó con tarjeta. Puede ser de tipo:
- `fixed` — gasto recurrente generado por un `RecurringExpense`. Difícilmente cambia (renta, internet, luz, streaming).
- `variable` — gasto registrado por un Member en un Period específico. No se replica.
- `msi` — gasto raíz de una compra a meses sin intereses. **No se muestra en el listado general de gastos** — solo existe como registro contable. Cada mes se genera automáticamente un Expense cuota con descripción `"MSI N/M"` en el Period correspondiente.

### Installment (tracker MSI)
Entidad que gestiona la vida de una compra MSI desde el Expense raíz. Contiene: monto total, número de cuotas, cuota actual e ID del Expense raíz. Cada vez que se abre un nuevo Period, verifica si hay installments activos y genera el Expense cuota correspondiente. Cuando `currentInstallment == totalInstallments`, el tracker se marca como completado y deja de generar cuotas.

### Income (futuro)
Ingreso declarado por un Member para un Period específico. Por ahora no se usa en cálculos — el settlement usa `Member.MonthlySalaryCents`. Preparado para cuando se necesiten ingresos variables por periodo (comisiones, horas extra, etc.) que sobreescriban el salario base.

### PeriodApproval
Voto de un Member sobre el cierre de un Period. Estados: `approved` u `objected`. El consenso es necesario para cerrar el Period; el owner del Household puede forzar el cierre ante objeciones.

### Balance (derivado)
Cálculo en tiempo real: `total_gastado_por_miembro - contribución_esperada`. Positivo = pagó de más (le deben). Negativo = pagó de menos (debe). Se liquida al cerrar el Period y no se acumula entre periodos.

### Card
Tarjeta registrada por un Member para registrar pagos. Es **personal** (cada miembro gestiona sus propias tarjetas). Se usa para:
- Seleccionar desde un dropdown al crear un Expense
- Diferenciar entre tarjetas (Banamex Oro vs BBVA Azul)
- Agrupar reportes por tarjeta

Cada Card tiene: banco, nombre visible y día de corte. La tarjeta **no se comparte** entre miembros — aunque un Expense pagado con ella sí es visible (el nombre de la tarjeta aparece en el gasto), la configuración y lista completa es privada de cada miembro.

Campos: `HouseholdID`, `OwnerMemberID`, `bankName`, `cardName`, `cutoffDay`.

### Category
Agrupación semántica de Expense. Puede ser predefinida (slugs: `rent`, `auto`, `streaming`, `food`, `personal`, `savings`, `other`) o personalizada por Household.
