# Historial de versiones

## 2.0.0

- Se simplifica el contrato JSON de observabilidad: los diagnósticos de ruta
  derivados y los contadores de conciliación redundantes dejan de formar parte
  de `HalcyonReport`. Los consumidores deben actualizar los tipos del cliente
  antes de aceptar esta versión.
- Se amplía la regresión de auditoría para capacidad, estado de ruta, deuda
  socializada y referencias de posición.
- Se conserva la semántica de liquidación de posiciones y la reproducción
  determinista de los escenarios existentes.

## 1.0.0

- Primera versión estable del motor de rutas, margen, financiación y
  liquidación, con CLI, cliente Node y matriz CI.
