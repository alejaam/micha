---
name: "test-db"
description: "Esta skill contiene scripts SQL para poblar la base de datos con datos de prueba. Útil para desarrollo local y pruebas de integración."
user-invocable: true
disable-model-invocation: true
metadata:
  - name: "reset.sh"
    description: "Script bash para reset completo de la DB (migraciones + seed)"
    author: "jalamar"
---

## Archivos

| Script | Descripción |
|---|---|
| `reset.sh` | Reset completo: borra todo y re-crea desde cero con migraciones + seed |
| `seed.sql` | Datos dummy base: users, household, members, cards, categories, expenses |
| `seed-flow-split.sql` | Flujo: split proporcional entre dos miembros con sueldos distintos |
| `seed-flow-installments.sql` | Flujo: gastos en MSI (meses sin intereses) con installments |
| `seed-flow-invitations.sql` | Flujo: invitación de miembro nuevo a un household existente |
| `clean.sql` | Limpia solo los datos (TRUNCATE) sin tocar estructura ni migraciones |
| `drop-recreate.sql` | DROP + CREATE de la DB (nuclear, para cuando necesitas empezar desde cero) |

---

## Uso rápido

```bash
# Desde la raíz del repo
cd backend

# Reset completo (migraciones + seed base)
bash scripts/test-db/reset.sh

# Solo limpiar datos (mantiene estructura)
psql $DATABASE_URL -f scripts/test-db/clean.sql

# Sembrar un flujo específico (después de reset base)
psql $DATABASE_URL -f scripts/test-db/seed-flow-installments.sql
```

---

## Variables de entorno

El script `reset.sh` usa `DATABASE_URL` del archivo `backend/.env`. Asegúrate de que apunte a tu DB **local/test** y no a staging.

```env
# backend/.env (local)
DATABASE_URL=postgres://micha:micha_dev_password@localhost:5432/micha_test?sslmode=disable
```

> ⚠️ **Nunca** ejecutes estos scripts contra staging. El `reset.sh` tiene una guardia que verifica que el host sea `localhost` o `127.0.0.1`.

---

## Usuarios de test pre-creados

| Email | Password | Rol |
|---|---|---|
| `alice@test.micha` | `test1234` | Admin del household |
| `bob@test.micha` | `test1234` | Miembro del household |
| `carol@test.micha` | `test1234` | Usuario sin household (para flows de invitación) |

Password hash generado con bcrypt cost 10: `$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LPVaLpPRfNi`

---

## Flujos cubiertos

### Flujo base (`seed.sql`)
- 2 usuarios con household `Casa Test`
- Settlement mode `proportional`, currency `MXN`
- 2 tarjetas (BBVA + Nu)
- 7 categorías por defecto
- 10 gastos variados (algunos con card, algunos en efectivo)

### Flujo Split Proporcional (`seed-flow-split.sql`)
- Alice gana 50,000 MXN/mes, Bob gana 30,000 MXN/mes
- Gastos compartidos con split calculado (62.5% / 37.5%)
- Útil para probar la pantalla de balance/liquidación

### Flujo MSI / Installments (`seed-flow-installments.sql`)
- Gasto de $24,000 MXN a 12 MSI
- 12 installments generadas (una por mes)
- Útil para probar vista de tarjeta y proyección

### Flujo Invitaciones (`seed-flow-invitations.sql`)
- Carol (sin household) recibe invitación de Alice
- Token de invitación pre-generado y activo
- Útil para probar el flow de onboarding / accept invite
