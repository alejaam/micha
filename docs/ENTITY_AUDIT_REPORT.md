# 📋 Auditoría Exhaustiva de Entidades de Dominio - Micha

**Fecha**: Marzo 22, 2026  
**Scope**: Búsqueda exhaustiva en test files, handlers, use cases y repositorios para identificar la API esperada vs la existente.

---

## 🎯 Hallazgo Principal

Se ha identificado una **BRECHA CRÍTICA** entre lo que los tests especifican y lo que está implementado:

- ✅ **EXPENSE**: ~95% compatible (solo naming issue menor)
- ⚠️ **CATEGORY**: ~50% compatible (constructor y métodos diferente)
- ❌ **MEMBER**: 30% compatible (estructura completa refactorizada)
- ❌ **HOUSEHOLD**: 20% compatible (arquitectura radicalmente diferente)

---

# 📊 MEMBER ENTITY

## Tabla de Campos (Attributes/DTO)

| Campo | Esperado | Existe | Tipo | Status |
|-------|----------|--------|------|--------|
| `ID` | ✓ SÍ | ✓ SÍ | `member.ID` | ✅ |
| `HouseholdID` | ✓ SÍ | ✓ SÍ | `string` | ✅ |
| `UserID` | ✓ SÍ | ✓ SÍ | `string` | ✅ |
| `Name` | ✓ SÍ | ❌ NO | `string` | ❌ FALTA |
| `Email` | ✓ SÍ | ❌ NO | `string` | ❌ FALTA |
| `MonthlySalaryCents` | ✓ SÍ | ❌ NO | `int64` | ❌ FALTA |
| `ContributionPct` | ❌ NO | ✓ SÍ | `float64` | ⚠️ EXTRA |
| `ValidFrom` | ❌ NO | ✓ SÍ | `time.Time` | ⚠️ EXTRA |
| `CreatedAt` | ✓ SÍ | ✓ SÍ | `time.Time` | ✅ |
| `UpdatedAt` | ✓ SÍ | ✓ SÍ | `time.Time` | ✅ |

### Nota sobre Struct
- **Esperado**: `member.Attributes` (referenciado en tests)
- **Existe**: `member.MemberAttributes` (en member.go)
- **Status**: ⚠️ Nombre diferente, estructura distinta

---

## Tabla de Métodos/Getters

| Método | Firma Esperada | Existe | Status |
|--------|---|---|---|
| `New()` | `New(id ID, hh string, name string, email string, salary int64, createdAt time.Time) (Member, error)` | ❌ | ❌ No coincide |
| `NewWithUserID()` | `NewWithUserID(id ID, hh string, name, email string, userID string, salary int64, createdAt time.Time) (Member, error)` | ❌ | ❌ FALTA |
| `NewFromAttributes()` | `NewFromAttributes(attrs Attributes) (Member, error)` | ✓ | ⚠️ Usa `MemberAttributes` |
| `ID()` | `ID() member.ID` | ✓ | ✅ |
| `HouseholdID()` | `HouseholdID() string` | ✓ | ✅ |
| `UserID()` | `UserID() string` | ✓ | ✅ |
| `Name()` | `Name() string` | ❌ | ❌ FALTA |
| `Email()` | `Email() string` | ❌ | ❌ FALTA |
| `MonthlySalaryCents()` | `MonthlySalaryCents() int64` | ❌ | ❌ FALTA |
| `ContributionPct()` | `ContributionPct() float64` | ✓ | ⚠️ No esperado |
| `ValidFrom()` | `ValidFrom() time.Time` | ✓ | ⚠️ No esperado |
| `CreatedAt()` | `CreatedAt() time.Time` | ✓ | ✅ |
| `UpdatedAt()` | `UpdatedAt() time.Time` | ✓ | ✅ |
| `LinkUser()` | `LinkUser(userID string)` | ❌ | ❌ FALTA |
| `UpdateProfile()` | `UpdateProfile(name, email string, salary int64) error` | ❌ | ❌ FALTA |
| `Attributes()` | `Attributes() Attributes` | ✓ | ⚠️ Retorna `MemberAttributes` |

---

## Tabla de Errores de Dominio

| Error | Esperado | Existe | Dónde | Status |
|-------|----------|--------|-------|--------|
| `member.ErrInvalidName` | ✓ | ❌ | En tests | ❌ FALTA |
| `member.ErrInvalidEmail` | ✓ | ❌ | En tests | ❌ FALTA |
| `member.ErrInvalidSalary` | ✓ | ❌ | En tests | ❌ FALTA |
| `shared.ErrInvalidID` | ✓ | ✓ | shared/errors.go | ✅ |
| `shared.ErrInvalidPercentage` | ✓ | ✓ | shared/errors.go | ✅ |

---

## Fuentes de Evidencia (MEMBER)

**Test File**: `internal/domain/member/member_test.go`
```go
member.NewFromAttributes(member.Attributes{
    ID:                 member.ID("m-1"),
    HouseholdID:        "hh-1",
    Name:               "Ale",
    Email:              "ale@mail.com",
    MonthlySalaryCents: 300000,
    CreatedAt:          time.Now(),
})
```

**Handler Usage**: `internal/adapters/http/member_handler.go`
```go
items = append(items, map[string]any{
    "name":                 attrs.Name,          // ← No existe
    "email":                attrs.Email,         // ← No existe
    "monthly_salary_cents": attrs.MonthlySalaryCents,  // ← No existe
})
```

**Use Case**: `internal/application/member/register_member.go`
```go
if strings.EqualFold(m.Email(), strings.TrimSpace(input.CallerEmail)) {  // ← Email() no existe
```

---

# 📊 HOUSEHOLD ENTITY

## Tabla de Campos (Attributes/DTO)

| Campo | Esperado | Existe | Tipo | Status |
|-------|----------|--------|------|--------|
| `ID` | ✓ SÍ | ✓ SÍ | `household.ID` | ✅ |
| `Name` | ✓ SÍ | ✓ SÍ | `string` | ✅ |
| `SettlementMode` | ✓ SÍ | ❌ NO | `SettlementMode` | ❌ FALTA |
| `Currency` | ✓ SÍ | ❌ NO | `string` | ❌ FALTA |
| `OwnerID` | ❌ NO | ✓ SÍ | `string` | ⚠️ EXTRA |
| `CycleType` | ❌ NO | ✓ SÍ | `CycleType` | ⚠️ EXTRA |
| `CreatedAt` | ✓ SÍ | ✓ SÍ | `time.Time` | ✅ |
| `UpdatedAt` | ✓ SÍ | ✓ SÍ | `time.Time` | ✅ |

### Nota sobre Struct
- **Esperado**: `household.Attributes` (referenciado en tests)
- **Existe**: `household.HouseholdAttributes` (en household.go)
- **Status**: ⚠️ Nombre diferente, estructura radicalmente distinta

---

## Tabla de Tipos/Constantes

| Tipo/Constante | Esperado | Existe | Valor | Status |
|---|---|---|---|---|
| `type SettlementMode string` | ✓ | ❌ | - | ❌ FALTA |
| `SettlementModeEqual` | ✓ | ❌ | `"equal"` | ❌ FALTA |
| `SettlementModeProportional` | ✓ | ❌ | `"proportional"` | ❌ FALTA |
| `type CycleType string` | ❌ | ✓ | - | ⚠️ EXTRA |
| `CycleTypeMonthly` | ❌ | ✓ | `"monthly"` | ⚠️ EXTRA |
| `CycleTypeBiweekly` | ❌ | ✓ | `"biweekly"` | ⚠️ EXTRA |
| `CycleTypeCustom` | ❌ | ✓ | `"custom"` | ⚠️ EXTRA |

---

## Tabla de Métodos/Getters

| Método | Firma Esperada | Existe | Status |
|--------|---|---|---|
| `NewFromAttributes()` | `NewFromAttributes(attrs Attributes) (Household, error)` | ✓ | ⚠️ Usa `HouseholdAttributes` |
| `ID()` | `ID() household.ID` | ✓ | ✅ |
| `Name()` | `Name() string` | ✓ | ✅ |
| `Currency()` | `Currency() string` | ❌ | ❌ FALTA |
| `SettlementMode()` | `SettlementMode() SettlementMode` | ❌ | ❌ FALTA |
| `OwnerID()` | `OwnerID() string` | ✓ | ⚠️ No esperado |
| `CycleType()` | `CycleType() CycleType` | ✓ | ⚠️ No esperado |
| `CreatedAt()` | `CreatedAt() time.Time` | ✓ | ✅ |
| `UpdatedAt()` | `UpdatedAt() time.Time` | ✓ | ✅ |
| `UpdateConfig()` | `UpdateConfig(name string, mode SettlementMode, currency string) error` | ❌ | ❌ FALTA |
| `Attributes()` | `Attributes() Attributes` | ✓ | ⚠️ Retorna `HouseholdAttributes` |

---

## Tabla de Errores de Dominio

| Error | Esperado | Existe | Fuente | Status |
|-------|----------|--------|--------|--------|
| `household.ErrInvalidName` | ✓ | ❌ | En tests | ❌ FALTA |
| `household.ErrInvalidSettlementMode` | ✓ | ❌ | En tests | ❌ FALTA |
| `household.ErrInvalidCurrency` | ✓ | ❌ | En tests | ❌ FALTA |
| `shared.ErrInvalidID` | ✓ | ✓ | shared/errors.go | ✅ |
| `shared.ErrInvalidStatus` | ✓ | ✓ | shared/errors.go | ✅ |

---

## Fuentes de Evidencia (HOUSEHOLD)

**Test File**: `internal/domain/household/household_test.go`
```go
household.NewFromAttributes(household.Attributes{
    ID:             household.ID("hh-1"),
    Name:           "Casa",
    SettlementMode: household.SettlementModeProportional,  // ← No existe
    Currency:       "mxn",                                  // ← No existe
    CreatedAt:      now,
})

if h.Currency() != "MXN" {  // ← Currency() no existe
    t.Errorf("Currency = %q; want %q", h.Currency(), "MXN")
}
```

**Repository Usage**: `internal/adapters/postgres/household_repository.go`
```go
string(attrs.ID), attrs.Name, string(attrs.SettlementMode), attrs.Currency,  // ← No existen campos
```

**Use Case**: `internal/application/household/update_household.go`
```go
if err := h.UpdateConfig(input.Name, input.SettlementMode, input.Currency); err != nil {  // ← No existe
```

---

# 📊 CATEGORY ENTITY

## Tabla de Campos (Attributes/DTO)

| Campo | Esperado | Existe | Tipo | Status |
|-------|----------|--------|------|--------|
| `ID` | ✓ | ✓ | `category.ID` | ✅ |
| `Name` | ✓ | ✓ | `string` | ✅ |
| `Slug` | ✓ | ❌ | `string` | ❌ FALTA |
| `CategoryType` | ❌ | ✓ | `CategoryType` | ⚠️ EXTRA |
| `HouseholdID` | ✓ | ✓ | `string` | ✅ |
| `CreatedAt` | ✓ | ✓ | `time.Time` | ✅ |
| `UpdatedAt` | ✓ | ✓ | `time.Time` | ✅ |

---

## Tabla de Constructor y Métodos

| Método | Firma Esperada | Firma Actual | Status |
|--------|---|---|---|
| `New()` | `New(id string, hh string, name string, slug string, createdAt time.Time) (Category, error)` | `New(id ID, name string, categoryType CategoryType, hh string, createdAt time.Time) (Category, error)` | ❌ Diferente |
| `Slug()` | `Slug() string` | ❌ No existe | ❌ FALTA |
| `IsDefault()` | `IsDefault() bool` | ❌ No existe | ❌ FALTA |
| `ID()` | ✓ | ✓ | ✅ |
| `Name()` | ✓ | ✓ | ✅ |
| `CategoryType()` | ❌ | ✓ | ⚠️ Extra |
| `HouseholdID()` | ✓ | ✓ | ✅ |
| `CreatedAt()` | ✓ | ✓ | ✅ |
| `UpdatedAt()` | ✓ | ✓ | ✅ |

---

## Validaciones Esperadas

| Regla | Esperado | Existe | Status |
|-------|----------|--------|--------|
| Slug pattern: `^[a-z0-9]+(?:-[a-z0-9]+)*$` | ✓ | ❌ | ❌ FALTA |
| Sin espacios en slug | ✓ | ❌ | ❌ FALTA |
| Sin uppercase en slug | ✓ | ❌ | ❌ FALTA |
| Sin special chars | ✓ | ❌ | ❌ FALTA |
| Sin leading/trailing hyphens | ✓ | ❌ | ❌ FALTA |
| Name no vacío | ✓ | ✓ | ✅ |

---

## Fuentes de Evidencia (CATEGORY)

**Test File**: `internal/domain/category/category_test.go`
```go
c, err := category.New("cat-1", "hh-1", "Gym", "gym", time.Now())
                                                        ^^^^← slug parameter

if c.Slug() != "gym" {  // ← Slug() method expected
    t.Errorf("Slug = %q; want %q", c.Slug(), "gym")
}

if c.IsDefault() {  // ← IsDefault() method expected
    t.Error("custom category should not be default")
}
```

---

# 📊 EXPENSE ENTITY

## Tabla de Campos (Attributes/DTO)

| Campo | Esperado | Existe | Tipo | Status |
|-------|----------|--------|------|--------|
| `ID` | ✓ | ✓ | `expense.ID` | ✅ |
| `MemberID` | ✓ | ✓ | `string` | ✅ |
| `PaidByMemberID` | ✓ | ⚠️ Alias | - | ⚠️ Legacy alias |
| `PeriodID` | ✓ | ✓ | `string` | ✅ |
| `CategoryID` | ✓ | ✓ | `string` | ✅ |
| `HouseholdID` | ✓ | ✓ | `string` | ✅ |
| `AmountCents` | ✓ | ✓ | `int64` | ✅ |
| `Description` | ✓ | ✓ | `string` | ✅ |
| `IsShared` | ✓ | ✓ | `bool` | ✅ |
| `Currency` | ✓ | ✓ | `string` | ✅ |
| `PaymentMethod` | ✓ | ✓ | `PaymentMethod` | ✅ |
| `ExpenseType` | ✓ | ✓ | `ExpenseType` | ✅ |
| `CardName` | ✓ | ✓ | `string` | ✅ |
| `CreatedAt` | ✓ | ✓ | `time.Time` | ✅ |
| `UpdatedAt` | ✓ | ✓ | `time.Time` | ✅ |
| `DeletedAt` | ✓ | ✓ | `*time.Time` | ✅ |

---

## Tabla de Métodos/Getters

| Método | Status |
|--------|--------|
| `New()` | ✅ |
| `NewFromAttributes()` | ✅ |
| `Patch()` | ✅ |
| `SoftDelete()` | ✅ |
| `ID()` | ✅ |
| `MemberID()` | ✅ |
| `PaidByMemberID()` (legacy alias) | ✅ |
| `PeriodID()` | ✅ |
| `CategoryID()` | ✅ |
| `HouseholdID()` | ✅ |
| `AmountCents()` | ✅ |
| `Description()` | ✅ |
| `IsShared()` | ✅ |
| `Currency()` | ✅ |
| `PaymentMethod()` | ✅ |
| `ExpenseType()` | ✅ |
| `CardName()` | ✅ |
| `CreatedAt()` | ✅ |
| `UpdatedAt()` | ✅ |
| `DeletedAt()` | ✅ |
| `Attributes()` | ✅ |

---

## Tabla de Tipos/Constantes

| Tipo | Valor | Status |
|-----|-------|--------|
| `PaymentMethodCash` | `"cash"` | ✅ |
| `PaymentMethodCard` | `"card"` | ✅ |
| `PaymentMethodTransfer` | `"transfer"` | ✅ |
| `PaymentMethodVoucher` | `"voucher"` | ✅ |
| `ExpenseTypeFixed` | `"fixed"` | ✅ |
| `ExpenseTypeVariable` | `"variable"` | ✅ |
| `ExpenseTypeMSI` | `"msi"` | ✅ |

---

## Tabla de Errores de Dominio

| Error | Status |
|-------|--------|
| `expense.ErrInvalidHouseholdID` | ✅ |
| `expense.ErrInvalidPaidByMemberID` | ✅ |
| `expense.ErrInvalidCurrency` | ✅ |
| `expense.ErrInvalidPaymentMethod` | ✅ |
| `expense.ErrInvalidExpenseType` | ✅ |
| `expense.ErrInvalidCategory` | ✅ |
| `shared.ErrInvalidMoney` | ✅ |
| `shared.ErrAlreadyDeleted` | ✅ |

---

# 🔍 ERRORES COMPARTIDOS DISPONIBLES

Desde `internal/domain/shared/errors.go`:

| Error | Existe | Usado Por |
|-------|--------|----------|
| `ErrInvalidMoney` | ✓ | EXPENSE, MEMBER, HOUSEHOLD |
| `ErrNotFound` | ✓ | General |
| `ErrAlreadyDeleted` | ✓ | EXPENSE |
| `ErrAlreadyExists` | ✓ | General |
| `ErrInvalidCredentials` | ✓ | AUTH |
| `ErrInvalidID` | ✓ | MEMBER, HOUSEHOLD, EXPENSE, CATEGORY |
| `ErrInvalidName` | ✓ | MEMBER, HOUSEHOLD, CATEGORY |
| `ErrInvalidPercentage` | ✓ | MEMBER |
| `ErrInvalidDateRange` | ✓ | PERIOD |
| `ErrInvalidStatus` | ✓ | HOUSEHOLD, CATEGORY |

---

# 📋 RESUMEN EJECUTIVO

## Matriz de Compatibilidad

```
EXPENSE:    ████████████████████░ 95% compatible ✅
CATEGORY:   ██████████░░░░░░░░░░░ 50% compatible ⚠️
MEMBER:     ███░░░░░░░░░░░░░░░░░░ 30% compatible ❌
HOUSEHOLD:  ██░░░░░░░░░░░░░░░░░░░ 20% compatible ❌
```

## Acciones Requeridas por Prioridad

### 🔴 CRÍTICO (Bloquea integración)
- [ ] **MEMBER**: Refactorizar estructura (Name, Email, MonthlySalaryCents + ContributionPct?)
- [ ] **HOUSEHOLD**: Refactorizar estructura (SettlementMode, Currency vs CycleType, OwnerID)
- [ ] **CATEGORY**: Decidir entre API basada en Slug vs CategoryType

### 🟠 ALTO (Requiere atención pronto)
- [ ] Crear métodos faltantes en MEMBER (LinkUser, UpdateProfile, NewWithUserID)
- [ ] Crear métodos faltantes en HOUSEHOLD (UpdateConfig)
- [ ] Crear métodos y validaciones en CATEGORY (Slug(), IsDefault(), slug validation)

### 🟡 MEDIO (Optimización)
- [ ] Crear errores específicos de dominio para MEMBER, HOUSEHOLD y CATEGORY
- [ ] Renombrar MemberAttributes → Attributes, HouseholdAttributes → Attributes

### 🟢 BAJO (Nice to have)
- [ ] Verificar si CategoryType debe convivir con Slug
- [ ] Documentar decisiones arquitectónicas

---

# 📌 Referencias Cruzadas

**Archivos de Código Actual**:
- [internal/domain/member/member.go](../../backend/internal/domain/member/member.go)
- [internal/domain/household/household.go](../../backend/internal/domain/household/household.go)
- [internal/domain/category/category.go](../../backend/internal/domain/category/category.go)
- [internal/domain/expense/expense.go](../../backend/internal/domain/expense/expense.go)
- [internal/domain/shared/errors.go](../../backend/internal/domain/shared/errors.go)

**Archivos de Test**:
- [internal/domain/member/member_test.go](../../backend/internal/domain/member/member_test.go)
- [internal/domain/household/household_test.go](../../backend/internal/domain/household/household_test.go)
- [internal/domain/category/category_test.go](../../backend/internal/domain/category/category_test.go)
- [internal/domain/expense/expense_test.go](../../backend/internal/domain/expense/expense_test.go)

**Adaptadores HTTP**:
- [internal/adapters/http/member_handler.go](../../backend/internal/adapters/http/member_handler.go)
- [internal/adapters/http/household_handler.go](../../backend/internal/adapters/http/household_handler.go)
- [internal/adapters/http/expense_handler.go](../../backend/internal/adapters/http/expense_handler.go)

**Casos de Uso**:
- [internal/application/member/](../../backend/internal/application/member/)
- [internal/application/household/](../../backend/internal/application/household/)
- [internal/application/expense/](../../backend/internal/application/expense/)
- [internal/application/category/](../../backend/internal/application/category/)

---

**Generado**: Marzo 22, 2026 | **Herramienta**: GitHub Copilot Entity Audit
