# Gobernanza

## Principios

Los cambios de parámetros deben ser explícitos, revisables, reversibles cuando sea posible y efectivos en una época conocida. Una propuesta incluye motivación, valores anterior y nuevo, simulaciones, límites y plan de reversión.

```mermaid
flowchart LR
    P["Propuesta"] --> S["Simulación"]
    S --> R["Revisión de riesgo"]
    R --> Q["Revisión técnica"]
    Q --> A["Aprobación"]
    A --> T["Espera temporal"]
    T --> E["Ejecución"]
    E --> C["Conciliación"]
```

## Parámetros gobernados

- objetivo y máximo de utilización;
- pendiente y límite de financiación;
- margen inicial y de mantenimiento;
- penalización de liquidación;
- participación y recuperación del seguro;
- estado de operadores, bóvedas y rutas;
- límites de cliente y política de publicación.

## Expediente de cambio

| Campo         | Contenido mínimo            |
| ------------- | --------------------------- |
| referencia    | identificador único         |
| alcance       | rutas, activos y operadores |
| valores       | anterior, nuevo y unidad    |
| justificación | señal cuantitativa          |
| simulación    | base, adversa y extrema     |
| activación    | época y responsables        |
| reversión     | condición y procedimiento   |

```mermaid
sequenceDiagram
    participant G as Proponente
    participant R as Riesgo
    participant M as Mantenedor
    participant O as Operaciones
    G->>R: expediente y simulaciones
    R-->>G: dictamen
    G->>M: cambio revisado
    M-->>O: artefacto y digest
    O->>O: activar en época acordada
    O-->>R: conciliación posterior
```

## Cambios de emergencia

Una emergencia permite pausar rutas o reducir límites, pero no omitir registro, revisión posterior ni conciliación. La autoridad temporal caduca y cualquier modificación permanente vuelve al flujo ordinario.

```mermaid
stateDiagram-v2
    [*] --> Normal
    Normal --> Emergency: señal confirmada
    Emergency --> Paused: contención
    Paused --> Review: estado estable
    Review --> Normal: aprobación ordinaria
    Review --> Retired: cierre definitivo
    Retired --> [*]
```

## Separación y evidencia

Proponente, aprobador y ejecutor deben ser identidades distintas. Se conserva commit, resultado de automatización, digest del artefacto, parámetros, hora efectiva y conciliación posterior. Una etiqueta publicada es inmutable; cualquier corrección crea una versión nueva.
