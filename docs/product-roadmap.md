# Micha - Roadmap de Producto

**Visión:** SaaS de finanzas compartidas para parejas, roomies y familias. Aplicación intuitiva para dividir gastos, calcular liquidaciones y analizar patrones de gasto.

---

## 🏠 **MÓDULO 1: Autenticación & Configuración Inicial** ✅
- [x] Registro e login
- [x] Autenticación multi-dispositivo
- [ ] Recuperación de contraseña
- [ ] Autenticación con OAuth (Google, Apple)
- [ ] Verificación por email
- [ ] Cambio y confirmación de contraseña

**Estado:** En progreso | **Prioridad:** Alta | **Sprint:** 0

---

## 👥 **MÓDULO 2: Gestión de Hogares & Miembros** ✅ (Parcial)
- [x] Crear hogar
- [x] Invitar miembros al hogar
- [x] Roles (admin, miembro, invitado)
- [ ] Editar información del hogar (nombre, divisa por defecto)
- [ ] Listar miembros activos
- [ ] Remover miembros
- [ ] Historial de cambios en miembros
- [ ] Configuración de permisos por rol

**Estado:** En progreso | **Prioridad:** Alta | **Sprint:** 0

---

## 💰 **MÓDULO 3: Registro de Gastos** ✅ (Parcial)
- [x] Crear gasto simple (qué, cuánto, quién pagó)
- [x] Asignar a qué miembros afecta el gasto
- [ ] Categorizar gasto (comida, alquiler, utilidades, entretenimiento, etc.)
- [ ] Etiquetas personalizadas
- [ ] Archivos adjuntos (tickets, fotos)
- [ ] Notas/descripción del gasto
- [ ] Editar/eliminar gastos
- [ ] Historial de cambios en gastos
- [ ] Duplicar gasto (para recurrentes de último momento)
- [ ] Gastos recurrentes (alquiler mensual, Netflix)

**Estado:** En progreso | **Prioridad:** Alta | **Sprint:** 1

---

## ➗ **MÓDULO 4: División de Gastos**
- [ ] División igual (50/50, 33/33/33)
- [ ] División por porcentaje personalizado
- [ ] División por monto específico
- [ ] División "quien pagó todo" (el que pagó se queda con el gasto)
- [ ] Gastos parciales (solo algunos miembros)
- [ ] Auto-cálculo de porcentajes

**Estado:** No iniciado | **Prioridad:** Alta | **Sprint:** 1

---

## 📊 **MÓDULO 5: Presupuestos & Control**
- [ ] Crear presupuestos por categoría (mensual/semanal)
- [ ] Presupuesto compartido vs. personal
- [ ] Alertas cuando se aproxima límite
- [ ] Visualizar gasto vs. presupuesto
- [ ] Análisis de overspending
- [ ] Recomendaciones basadas en patrones

**Estado:** No iniciado | **Prioridad:** Media | **Sprint:** 2

---

## 🧮 **MÓDULO 6: Liquidaciones & Deudas** ✅
- [x] Calcular quién debe a quién
- [x] Algoritmo de liquidación (minimizar transacciones)
- [ ] Generar resumen de deudas
- [ ] Marcar pagos como realizados
- [ ] Historial de liquidaciones
- [ ] Sugerir forma óptima de pagar
- [ ] Recordatorios de deudas pendientes

**Estado:** En progreso | **Prioridad:** Alta | **Sprint:** 1

---

## 📈 **MÓDULO 7: Reportes & Análisis**
- [ ] Resumen mensual (gastos totales, por persona, por categoría)
- [ ] Gráficos de gastos (pie, barras, línea temporal)
- [ ] Tendencias (mes a mes, año a año)
- [ ] Top categorías de gasto
- [ ] Distribución de gastos por miembro
- [ ] Balances acumulados (quién ha pagado más/menos)
- [ ] Exportar reportes (PDF, CSV, Excel)
- [ ] Dashboard interactivo con KPIs

**Estado:** No iniciado | **Prioridad:** Media | **Sprint:** 2

---

## 🔔 **MÓDULO 8: Notificaciones & Reminders**
- [ ] Notificaciones cuando se agrega gasto
- [ ] Reminders de deudas pendientes
- [ ] Notificaciones de liquidaciones sugeridas
- [ ] Resumen diario/semanal/mensual
- [ ] Alertas de presupuesto excedido
- [ ] Email digest

**Estado:** No iniciado | **Prioridad:** Media | **Sprint:** 3

---

## 🏘️ **MÓDULO 9: Múltiples Hogares & Contextos**
- [ ] Cambiar entre hogares (Ej: "Departamento" vs "Viaje por Europa")
- [ ] Gastos diferentes para cada hogar
- [ ] Miembros pueden estar en múltiples hogares
- [ ] Presupuestos independientes por hogar
- [ ] Historial separado por hogar

**Estado:** No iniciado | **Prioridad:** Baja | **Sprint:** 3

---

## 🔗 **MÓDULO 10: Integraciones & Importación**
- [ ] Conectar tarjeta de crédito (automatizar captura de gastos)
- [ ] Sincronizar con Spotify, Netflix (detectar gastos compartidos)
- [ ] Importar CSV de gastos
- [ ] API para integraciones externas
- [ ] Webhook para eventos

**Estado:** No iniciado | **Prioridad:** Baja | **Sprint:** 4

---

## 🔒 **MÓDULO 11: Seguridad & Privacidad**
- [ ] Encriptación de datos sensibles
- [ ] Autenticación de dos factores (2FA)
- [ ] Control de acceso por rol
- [ ] Auditoría de accesos
- [ ] Exportar mis datos (GDPR)
- [ ] Eliminar cuenta

**Estado:** No iniciado | **Prioridad:** Alta | **Sprint:** 3

---

## 🎨 **MÓDULO 12: Experiencia de Usuario**
- [ ] Interfaz mobile-first
- [ ] Modo offline (sincronizar cuando hay conexión)
- [ ] Búsqueda y filtros avanzados
- [ ] Temas oscuro/claro
- [ ] Internacionalización (ES, EN, PT)
- [ ] Accesibilidad (WCAG)
- [ ] Animaciones fluidas

**Estado:** En progreso | **Prioridad:** Alta | **Sprint:** 1+

---

## ⚙️ **MÓDULO 13: Configuración & Preferencias**
- [ ] Cambiar moneda del hogar
- [ ] Formato de fecha/hora
- [ ] Zona horaria
- [ ] Notificaciones personalizables
- [ ] Privacidad de datos

**Estado:** No iniciado | **Prioridad:** Media | **Sprint:** 2

---

## 💳 **MÓDULO 14: Monetización (Opcional)**
- [ ] Plan Free (hasta 3 miembros, 100 transacciones)
- [ ] Plan Pro ($2.99/mes) - miembros ilimitados, reportes, integraciones
- [ ] Plan Family ($4.99/mes) - todo + hasta 5 hogares

**Estado:** No iniciado | **Prioridad:** Baja | **Sprint:** 5+

---

## 📋 **Fases de Desarrollo Recomendadas**

### **Fase 1 (MVP - Sprint 0)**
Enfoque: Lo mínimo para que un usuario pueda dividir gastos y liquidarlos
- ✅ Autenticación
- ✅ Gestión de hogares & miembros (básico)
- ✅ Registro de gastos simple
- ➗ División igual de gastos
- ✅ Cálculo de liquidaciones

**Tiempo estimado:** 1-2 semanas

---

### **Fase 2 (v1.0 - Sprint 1)**
Enfoque: Mejorar la experiencia y agregar funcionalidades básicas
- Categorización de gastos
- Presupuestos básicos
- Reportes simples
- Historial de gastos
- Editar/eliminar gastos

**Tiempo estimado:** 2 semanas

---

### **Fase 3 (v1.5 - Sprint 2)**
Enfoque: Profundizar en análisis y automatización
- Gastos recurrentes
- Notificaciones
- Múltiples hogares
- Reportes avanzados (gráficos, tendencias)
- Alertas de presupuesto

**Tiempo estimado:** 2 semanas

---

### **Fase 4 (v2.0 - Sprint 3+)**
Enfoque: Escalabilidad y features premium
- Integraciones bancarias
- Integraciones con apps (Spotify, Netflix)
- Analytics profundo
- Monetización (planes)
- 2FA y seguridad avanzada

**Tiempo estimado:** 4+ semanas

---

## 🎯 **Notas de Iteración**

- **Sprint 0 (Actual):** Pulir autenticación, hogares y miembros. Agregar división igual de gastos.
- **Sprint 1:** Agregar categorías y reportes simples.
- **Sprint 2:** Profundizar en UX mobile y reportes avanzados.
- **Retroalimentación:** Recopilar datos de usuarios beta antes de monetizar.

---

## 🔄 **Última actualización**
Fecha: 2026-03-14 | Versión: 0.1
