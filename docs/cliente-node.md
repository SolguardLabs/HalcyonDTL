# Cliente Node

## Propósito

El paquete exporta `HalcyonClient`, `HalcyonClientError`, `assertScenarioName` y `validateReport`. El adaptador mantiene pequeña la superficie de integración y deja la lógica económica en Go.

```mermaid
flowchart LR
    APP["Aplicación"] --> CLIENT["HalcyonClient"]
    CLIENT --> PROC["Proceso Go"]
    PROC --> JSON["Reporte JSON"]
    JSON --> VALID["Validación"]
    VALID --> FREEZE["deepFreeze"]
    FREEZE --> APP
```

## Uso

```js
import { HalcyonClient } from "halcyondtl";

const client = new HalcyonClient({
  binary: process.env.HALCYON_BIN,
  timeout: 5_000,
  maxBuffer: 2 * 1024 * 1024,
});

for (const name of client.listScenarios()) {
  const report = client.runScenario(name);
  console.log(name, report.state_digest);
}
```

## Frontera de proceso

```mermaid
sequenceDiagram
    participant A as Aplicación
    participant C as Cliente
    participant P as Proceso
    A->>C: runScenario nombre
    C->>C: validar nombre
    C->>P: argv sin shell
    P-->>C: JSON y estado
    C->>C: validar identidad y digest
    C->>C: congelar objeto
    C-->>A: reporte
```

No se interpolan argumentos en comandos. El nombre debe comenzar por una letra minúscula y contener sólo letras minúsculas, números o guiones, con un máximo de 48 caracteres. El proceso se oculta en Windows y se termina al superar el plazo.

## Errores

`HalcyonClientError` separa ausencia del binario, error de ejecución, rechazo del motor, JSON mal formado e incumplimiento del contrato. `details` se congela y limita `stderr` a 4096 caracteres para evitar transportar salidas sin límite.

```mermaid
flowchart TD
    R["Solicitud"] --> N{"Nombre válido"}
    N -- no --> T["TypeError"]
    N -- sí --> E{"Binario existe"}
    E -- no --> H["HalcyonClientError"]
    E -- sí --> X["Ejecutar"]
    X --> S{"Estado cero"}
    S -- no --> H
    S -- sí --> J{"JSON válido"}
    J -- no --> H
    J -- sí --> O["Reporte inmutable"]
```

## Integración segura

- Fija el binario mediante ruta absoluta en servicios.
- No aceptes el nombre del escenario desde una fuente sin validar.
- Conserva límites de tiempo y memoria.
- Comprueba la versión del paquete junto al commit del binario.
- Registra el digest, no el reporte completo, cuando haya datos operativos sensibles.
