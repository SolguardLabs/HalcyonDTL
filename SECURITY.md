# Security Policy

HalcyonDTL es un laboratorio CTF. El codigo contiene un fallo economico
intencional y no debe usarse como implementacion productiva.

## Alcance

El alcance de auditoria incluye:

- acumuladores de funding por ruta;
- snapshots de cuenta;
- apertura y cierre de posiciones;
- liquidaciones;
- socializacion de deuda y seguro del pool;
- reportes JSON emitidos por la CLI.

## Fuera De Alcance

- disponibilidad de la CLI local;
- configuracion de CI;
- dependencias de desarrollo;
- ausencia de criptografia real en los fixtures.

## Reportes

En un entorno CTF, los reportes deben incluir:

- precondiciones economicas;
- secuencia de transiciones;
- impacto cuantificado;
- por que las invariantes publicas no detectan el fallo;
- mitigacion directa.

