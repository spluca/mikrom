# Migración de Ansible a gRPC

Este documento describe los cambios realizados para eliminar completamente Ansible del proyecto mikrom y reemplazarlo con comunicación gRPC hacia firecracker-agent.

## Resumen de Cambios

### 1. Nuevos Componentes

#### **api/proto/firecracker/v1/**

Directorio que contiene los archivos Protocol Buffers para la comunicación gRPC:

- `firecracker.proto` - Definición de servicios y mensajes gRPC
- `firecracker.pb.go` - Código Go generado para los mensajes
- `firecracker_grpc.pb.go` - Código Go generado para el cliente/servidor gRPC

#### **pkg/grpcclient/client.go**

Nuevo cliente gRPC que reemplaza el antiguo wrapper de Ansible:

- `NewClient(addr string)` - Crea conexión gRPC con firecracker-agent
- `CreateVM(ctx, params)` - Crea una VM vía gRPC
- `StartVM(ctx, vmID)` - Inicia una VM vía gRPC
- `StopVM(ctx, vmID, force)` - Detiene una VM vía gRPC
- `DeleteVM(ctx, vmID)` - Elimina una VM vía gRPC
- `GetVM(ctx, vmID)` - Obtiene información de una VM
- `HealthCheck(ctx)` - Verifica la salud del firecracker-agent

### 2. Archivos Modificados

#### **config/config.go**

**Cambios:**

- ❌ Eliminado: `FirecrackerDeployPath` (ruta a playbooks de Ansible)
- ❌ Eliminado: `FirecrackerDefaultHost` (host de Ansible)
- ✅ Agregado: `FirecrackerAgentAddr` (dirección gRPC del firecracker-agent)

**Nueva configuración:**

```go
type Config struct {
    // ...
    FirecrackerAgentAddr string  // Default: "localhost:50051"
    // ...
}
```

#### **pkg/worker/handlers.go**

**Cambios:**

- Reemplazado `*firecracker.Client` por `*grpcclient.Client`
- Eliminado el campo `firecrackerPath`
- Actualizado `NewTaskHandler()` para recibir el cliente gRPC

**Handlers actualizados:**

- `HandleCreateVM()` - Ahora usa `grpcClient.CreateVM()`
- `HandleStartVM()` - Ahora usa `grpcClient.StartVM()`
- `HandleStopVM()` - Ahora usa `grpcClient.StopVM()`
- `HandleRestartVM()` - Ahora usa `grpcClient.StopVM()` + `grpcClient.StartVM()`
- `HandleDeleteVM()` - Ahora usa `grpcClient.DeleteVM()`

#### **cmd/worker/main.go**

**Cambios:**

- Reemplazado import de `pkg/firecracker` por `pkg/grpcclient`
- Eliminada inicialización del cliente Ansible
- Agregada conexión gRPC con firecracker-agent
- Agregado health check del firecracker-agent

**Antes:**

```go
fcClient := firecracker.NewClient(cfg.FirecrackerDeployPath, cfg.FirecrackerDefaultHost)
taskHandler := worker.NewTaskHandler(db.DB, vmRepo, ipPoolRepo, fcClient, cfg.FirecrackerDeployPath)
```

**Después:**

```go
grpcClient, err := grpcclient.NewClient(cfg.FirecrackerAgentAddr)
defer grpcClient.Close()
taskHandler := worker.NewTaskHandler(db.DB, vmRepo, ipPoolRepo, grpcClient)
```

#### **.env.example**

**Cambios:**

- ❌ Eliminado: `FIRECRACKER_DEPLOY_PATH`
- ❌ Eliminado: `FIRECRACKER_DEFAULT_HOST`
- ✅ Agregado: `FIRECRACKER_AGENT_ADDR=localhost:50051`

#### **README.md**

**Cambios:**

- Actualizada descripción del proyecto
- Eliminadas referencias a Ansible en Features
- Actualizado Prerequisites (Ansible → firecracker-agent)
- Actualizado diagrama de arquitectura
- Actualizada sección de configuración
- Actualizada estructura del proyecto

### 3. Componentes Eliminados

#### **pkg/firecracker/** (directorio completo)

Eliminado el wrapper de Ansible que contenía:

- `client.go` - Cliente que ejecutaba playbooks de Ansible
- Métodos: `CreateVM()`, `StartVM()`, `StopVM()`, `CleanupVM()`

### 4. Dependencias Agregadas

Nuevas dependencias en `go.mod`:

```
google.golang.org/grpc v1.78.0
google.golang.org/protobuf v1.36.11
google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda
```

## Configuración Requerida

### Variables de Entorno

Agregar a tu archivo `.env`:

```env
# Firecracker Agent Configuration (gRPC)
FIRECRACKER_AGENT_ADDR=localhost:50051
```

Si firecracker-agent corre en otro servidor:

```env
FIRECRACKER_AGENT_ADDR=192.168.1.100:50051
```

### firecracker-agent

Asegúrate de que firecracker-agent esté corriendo antes de iniciar el worker:

```bash
# En el servidor donde corre firecracker-agent
cd /path/to/firecracker-agent
./bin/fc-agent --config configs/agent.yaml
```

## Cómo Usar

### 1. Iniciar firecracker-agent

```bash
cd /path/to/firecracker-agent
./bin/fc-agent --config configs/agent.yaml
```

Por defecto, escucha en `localhost:50051`

### 2. Configurar mikrom

Editar `.env`:

```env
FIRECRACKER_AGENT_ADDR=localhost:50051  # o la IP del servidor
```

### 3. Iniciar mikrom

```bash
# Terminal 1: API
go run cmd/api/main.go

# Terminal 2: Worker
go run cmd/worker/main.go
```

El worker se conectará automáticamente al firecracker-agent y verificará su salud.

## Ventajas de gRPC vs Ansible

1. **Rendimiento**: gRPC es mucho más rápido que ejecutar playbooks de Ansible
   - Ansible: ~5-10 segundos por operación
   - gRPC: < 500ms por operación

2. **Comunicación binaria**: Protocol Buffers es más eficiente que JSON/YAML

3. **Tipado fuerte**: Las interfaces gRPC están definidas en `.proto`

4. **Streaming**: Posibilidad de streams bidireccionales (para eventos en tiempo real)

5. **Sin dependencias externas**: No requiere Ansible instalado en el servidor

6. **Mejor manejo de errores**: Códigos de error gRPC estándar

## Verificación

Para verificar que todo funciona correctamente:

```bash
# 1. Build
go build -o bin/api ./cmd/api
go build -o bin/worker ./cmd/worker

# 2. Test
go test ./...

# 3. Verificar conexión gRPC
# El worker debería mostrar:
# [timestamp] Connecting to firecracker-agent at localhost:50051...
# [timestamp] firecracker-agent health check passed
```

## Troubleshooting

### Error: "failed to connect to firecracker-agent"

**Solución:**

1. Verificar que firecracker-agent esté corriendo: `ps aux | grep fc-agent`
2. Verificar el puerto: `netstat -tulpn | grep 50051`
3. Revisar la variable `FIRECRACKER_AGENT_ADDR` en `.env`

### Error: "firecracker-agent health check failed"

**Solución:**

1. Verificar logs de firecracker-agent
2. Intentar curl manual al health endpoint
3. Revisar firewall/networking entre mikrom y firecracker-agent

### Worker inicia pero las VMs no se crean

**Solución:**

1. Revisar logs del worker: errores de gRPC
2. Revisar logs de firecracker-agent: errores de Firecracker
3. Verificar que Firecracker esté instalado en el servidor del agent

## Notas Adicionales

- Los tests no requieren modificación ya que no hacían mocking del cliente de Ansible
- La API REST no cambió, los endpoints siguen siendo los mismos
- La migración es transparente para los usuarios de la API
- No se requiere migración de datos en la base de datos

## Siguiente Fase

Para completar la integración:

1. Implementar la Fase 2 de firecracker-agent (integración real con Firecracker)
2. Agregar soporte para eventos en tiempo real vía gRPC streaming
3. Implementar TLS/mTLS para comunicación segura entre mikrom y firecracker-agent
