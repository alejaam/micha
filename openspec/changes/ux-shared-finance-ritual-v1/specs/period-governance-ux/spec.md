# period-governance-ux Specification

## Purpose

Definir cómo la UX comunica y aplica la gobernanza del periodo (`open`, `review`, `closed`) para evitar acciones incorrectas y facilitar decisiones en hogares compartidos.

## Requirements

### Requirement: Estado de periodo con bloqueo explícito de acciones

La aplicación **MUST** reflejar el estado de periodo y bloquear acciones mutables cuando el periodo no esté abierto.

#### Scenario: Periodo abierto permite mutaciones

- GIVEN un hogar en estado `open`
- WHEN una persona intenta crear, editar o eliminar gastos
- THEN el sistema permite la acción
- AND muestra estado de periodo abierto en la interfaz

#### Scenario: Periodo en review/closed bloquea mutaciones

- GIVEN un hogar en estado `review` o `closed`
- WHEN una persona intenta crear, editar o eliminar gastos
- THEN el sistema bloquea la acción
- AND muestra el motivo del bloqueo en mensaje visible y consistente

### Requirement: Historial y datos provisionales transparentes

La aplicación **SHALL** indicar de forma explícita cuándo un dato histórico es provisional y cuál es su nivel de confiabilidad.

#### Scenario: Indicador de historial provisional

- GIVEN un periodo histórico con fallback o datos incompletos
- WHEN la persona consulta historial y tendencias
- THEN el sistema marca el contenido como provisional
- AND explica por qué se muestra provisional

#### Scenario: Próximo paso recomendado ante datos provisionales

- GIVEN datos históricos provisionales visibles
- WHEN la persona usuaria revisa el panel histórico
- THEN el sistema muestra la acción sugerida para obtener datos definitivos
- AND mantiene disponible el contexto suficiente para conciliación sin ambigüedad
