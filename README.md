# ResetPassJ

Secure password reset module for Go. Uses **bbolt** for storage and **gomail** for sending emails.

---

## Instalación

Instala las dependencias necesarias:

```bash
go get gopkg.in/gomail.v2
go get go.etcd.io/bbolt
```

# Requerimientos Funcionales - Módulo de Envío y Validación de Códigos

| Código    | Requerimiento                         | Descripción                                                                                                                                            | Prioridad |
| --------- | ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ | --------- |
| **RF-01** | Recepción de correo                   | El sistema deberá recibir un correo electrónico para iniciar el proceso de envío de código.                                                            | Alta      |
| **RF-02** | Generación de código                  | El sistema deberá generar un código único y aleatorio.                                                                                                 | Alta      |
| **RF-03** | Envío de código por correo            | El sistema deberá enviar el código generado al correo proporcionado.                                                                                   | Alta      |
| **RF-04** | Registro en BBolt                     | El sistema deberá almacenar en BBolt el correo, el código generado, la fecha/hora y el estado (usado / no usado).                                      | Alta      |
| **RF-05** | Marcado de uso                        | Cuando el código sea validado, el sistema deberá marcarlo como usado para impedir su reutilización.                                                    | Alta      |
| **RF-06** | Límite de intentos configurable       | El sistema deberá permitir configurar el número máximo de intentos por correo en un periodo. Por defecto, el valor será **3 intentos**.                | Alta      |
| **RF-07** | Bloqueo por alcanzar el límite        | Si un correo alcanza el número máximo de intentos configurado, el sistema deberá bloquear nuevos envíos durante un periodo de bloqueo.                 | Alta      |
| **RF-08** | Duración de bloqueo configurable      | El sistema deberá permitir configurar la duración del bloqueo. Por defecto será **24 horas**, pero podrá establecerse en horas o días según necesidad. | Alta      |
| **RF-09** | Eliminación de intentos no bloqueados | Si el usuario no alcanza el número máximo de intentos configurado, los intentos deberán eliminarse automáticamente después del día en curso.           | Alta      |
| **RF-10** | Reset post-bloqueo                    | Una vez transcurrido el periodo de bloqueo configurado, el usuario podrá solicitar nuevamente un código con un contador de intentos reiniciado.        | Alta      |
