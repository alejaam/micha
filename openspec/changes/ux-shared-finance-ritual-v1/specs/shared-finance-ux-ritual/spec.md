# shared-finance-ux-ritual Specification

## Purpose

Definir la experiencia para hogares compartidos (pareja/roomies) con un flujo claro de registrar gastos, conciliar y cerrar periodo, usando lenguaje consistente y reglas explícitas.

## Requirements

### Requirement: Consistencia de lenguaje y flujo colaborativo

La interfaz **MUST** usar lenguaje consistente en español para acciones, estados y feedback en las vistas core de hogar compartido.

#### Scenario: Flujo principal con copy consistente

- GIVEN una persona usuaria autenticada dentro de un hogar
- WHEN navega entre Dashboard, Gastos, Balances y Reglas
- THEN las etiquetas de navegación, CTAs y mensajes de estado se muestran en español consistente
- AND el flujo comunica registrar → conciliar → cerrar periodo

#### Scenario: Feedback de error accionable

- GIVEN una acción bloqueada o inválida en flujo de gastos
- WHEN el sistema muestra feedback al usuario
- THEN el mensaje describe el motivo en lenguaje claro
- AND incluye la siguiente acción recomendada cuando aplique

### Requirement: Guardrails en captura de gastos compartidos

El sistema **MUST** guiar la captura de gastos conforme a reglas de negocio de hogar compartido.

#### Scenario: Tipo fixed restringido al setup dedicado

- GIVEN una persona usuaria en el modal/formulario general de alta de gasto
- WHEN intenta registrar un gasto de tipo fixed
- THEN el sistema impide esa opción en el flujo general
- AND dirige a la sección de setup de gastos fijos del hogar

#### Scenario: Claridad de “quién pagó”

- GIVEN una persona usuaria en creación de gasto
- WHEN selecciona el campo “pagado por”
- THEN el sistema muestra miembros elegibles según reglas del hogar/rol
- AND deja explícito si registra para sí o en nombre de otra persona permitida
