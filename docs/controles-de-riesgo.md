# Controles de riesgo

## Capas

El control se distribuye entre admisión, capacidad de ruta, margen, financiación, liquidación, seguro e invariantes. Ninguna señal aislada se considera suficiente.

```mermaid
flowchart TD
    I["Entrada"] --> A["Política de admisión"]
    A --> C["Capacidad de ruta"]
    C --> M["Margen inicial"]
    M --> F["Financiación por época"]
    F --> Q["Margen de mantenimiento"]
    Q --> L["Liquidación"]
    L --> S["Seguro"]
    S --> D["Deuda socializada"]
```

## Límites recomendados

| Señal                |  Umbral de aviso | Acción                   |
| -------------------- | ---------------: | ------------------------ |
| utilización          |             90 % | limitar nuevas aperturas |
| utilización          |             96 % | pausar capacidad         |
| margen               | 120 % del mínimo | revisión frecuente       |
| antigüedad de precio |         2 épocas | rechazar transición      |
| déficit residual     |   mayor que cero | incidente crítico        |
| digest inesperado    | cualquier cambio | detener consumo          |

Los valores concretos deben ajustarse a liquidez, volatilidad y latencia del dominio operativo.

```mermaid
quadrantChart
    title "Priorización de señales"
    x-axis "Impacto bajo" --> "Impacto alto"
    y-axis "Probabilidad baja" --> "Probabilidad alta"
    quadrant-1 "Acción inmediata"
    quadrant-2 "Reducir exposición"
    quadrant-3 "Observar"
    quadrant-4 "Preparar contingencia"
    "Precio caducado": [0.80, 0.75]
    "Capacidad al límite": [0.72, 0.84]
    "Operador degradado": [0.55, 0.58]
    "Diferencia de digest": [0.92, 0.35]
```

## Invariantes

El reporte comprueba cuentas y bóvedas no negativas, utilización dentro del máximo, enlaces de posición, correspondencia entre registros y rutas, cursores activos válidos y ausencia de posiciones cerradas en cuentas.

```mermaid
flowchart LR
    ST["Estado"] --> I1["Saldos no negativos"]
    ST --> I2["Capacidad acotada"]
    ST --> I3["Enlaces completos"]
    ST --> I4["Historial coherente"]
    I1 --> G["Puerta de aceptación"]
    I2 --> G
    I3 --> G
    I4 --> G
```

## Pruebas de estrés

Cada revisión de parámetros debe incluir escenario base, concentración en una ruta, recorte de valoración, coste de liquidación elevado, recuperación parcial del seguro y choque simultáneo en cartera. Se archivan entradas y resultados, no sólo el nivel final.

## Separación de funciones

La persona que propone parámetros no debe aprobar su publicación. Operaciones controla pausas; riesgo valida límites; ingeniería mantiene el motor; y seguridad verifica integridad y procedencia del artefacto.
