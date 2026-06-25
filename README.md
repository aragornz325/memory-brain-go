# Memory Brain Backend (Go Edition)

Esta es una recreación/migración profesional del backend y herramienta CLI de `memory-brain` desde NestJS a Go (1.22+), utilizando Clean Architecture pragmática, base de datos PostgreSQL con `pgvector` para indexación de embeddings semánticos, y Ollama como proveedor de embeddings local.

## Arquitectura

El backend está estructurado siguiendo principios de Clean Architecture sin sobreingeniería:

- **`cmd/`**: Puntos de entrada para el servidor HTTP (`server/main.go`) y la herramienta CLI (`cli/main.go`).
- **`internal/domain/`**: Entidades del dominio (`MemoryItem`, `Workspace`, etc.) y errores tipados del dominio (`errors.go`). Esta capa es 100% pura y no tiene dependencias de infraestructura ni de frameworks de base de datos o HTTP.
- **`internal/config/`**: Configuración estructurada agrupada (`HTTP`, `Database`, `Ollama`, `Auth`) cargada mediante `godotenv` y variables de entorno.
- **`internal/database/`**: Pool de conexión de base de datos (`pgxpool`).
- **`internal/repository/`**: Capa de persistencia SQL sobre PostgreSQL usando `pgx`. Intercepta los errores de base de datos (como `pgx.ErrNoRows`) y los mapea a errores de dominio correspondientes.
- **`internal/service/`**: Lógica de negocio y consumo de APIs de embedding (Ollama). Aquí se definen las interfaces de repositorio de datos (`interfaces.go`), siguiendo el principio idiomático de que **las interfaces pertenecen al consumidor**.
- **`internal/http/`**: Router HTTP (Chi), middlewares globales (logging con `slog` y límite de tamaño de request con `http.MaxBytesReader` para mitigar ataques DoS), autenticación por API Key y controladores (handlers). Los controladores traducen los errores de dominio del service a códigos de estado HTTP semánticos (404, 401, 409, etc.).
- **`internal/cli/`**: Estructura de comandos y flags del CLI usando `spf13/cobra`.

---

## Estructura y Organización de Memorias (Workspace → Project → Memory)

El conocimiento registrado se estructura de forma jerárquica en tres niveles para facilitar su búsqueda y reutilización en el futuro:

```
Workspace (ej. personal-lab)
  ├── Project A (ej. titan)
  │     ├── Memory 1
  │     └── Memory 2
  └── Project B (ej. memory-brain)
        ├── Memory 3
        └── Memory 4
```

### Convenciones de Proyectos Oficiales
Las memorias deben clasificarse estrictamente según el proyecto en el que este conocimiento será útil en el futuro, en lugar de utilizar un único proyecto genérico.

1. **`memory-brain`**: Reservado exclusivamente para la arquitectura de este backend de Go, API HTTP, diseño del CLI, base de datos PostgreSQL, configuración del pool de conexiones y búsquedas vectoriales `pgvector`.
2. **`titan`**: Reservado únicamente para la infraestructura del laboratorio de inteligencia artificial local (configuración de Ollama, GPUs, drivers de Nvidia, integración con VSCode/Continue, Docker y hardware).
3. **Otros Proyectos**: Cada aplicación, microservicio o caso de uso independiente debe poseer su propio `ProjectSlug` representativo.

### Creación Automática de Proyectos (Fase 4 & 5)
Para agilizar el flujo de desarrollo, el backend cuenta con **creación automática de proyectos**.
* **Flujo**: Al intentar registrar una memoria (ya sea vía CLI o API HTTP) con un `ProjectSlug` que no existe en el `Workspace` especificado, el servidor Go interceptará el error de ausencia del proyecto y lo **creará automáticamente y de forma transaccional**.
* **Básicos**: Si el **Workspace** no existe, el servidor retornará un error HTTP `404 Not Found` de forma transparente. Si el Workspace sí existe, el proyecto se creará y la memoria se enlazará en el momento del guardado sin requerir interacción manual por parte del usuario.

---

## Tecnologías Utilizadas

- **Lenguaje**: Go Moderno (1.22)
- **HTTP Router**: [go-chi/chi](https://github.com/go-chi/chi) (ligero y compatible con `net/http` nativo)
- **Base de Datos**: PostgreSQL con la extensión `pgvector`
- **Driver de BD**: [jackc/pgx/v5](https://github.com/jackc/pgx) (alto rendimiento y soporte nativo para arrays y JSONB)
- **Carga de Entorno**: [joho/godotenv](https://github.com/joho/godotenv)
- **Logging**: `log/slog` (estructurado en JSON en la biblioteca estándar)
- **Herramienta CLI**: [spf13/cobra](https://github.com/spf13/cobra)
- **Herramienta de Migración**: [Goose](https://github.com/pressly/goose)

---

## Requisitos de Entorno

- **Go 1.22+** instalado.
- **Docker y Docker Compose** para levantar la base de datos PostgreSQL con `pgvector`.
- **Ollama** ejecutándose localmente o en la red con el modelo `nomic-embed-text` descargado (`ollama pull nomic-embed-text`).

---

## Preparación del Entorno Local

1. **Copiar variables de entorno**:
   ```bash
   cp .env.example .env
   ```
2. **Levantar base de datos con pgvector**:
   ```bash
   make run-db
   ```
3. **Ejecutar Migraciones**:
   Instalar `goose` si no se tiene (`go install github.com/pressly/goose/v3/cmd/goose@latest`), y ejecutar:
   ```bash
   goose -dir migrations postgres "postgres://memory_brain:memory_brain@localhost:5432/memory_brain?sslmode=disable" up
   ```

---

## Ejecución y Construcción

- **Construir binarios del Servidor y CLI**:
  ```bash
  make build
  # Los binarios se compilarán en `bin/server` y `bin/memory`
  ```
- **Iniciar Servidor HTTP de la API**:
  ```bash
  make run-server
  ```
- **Ejecutar Pruebas**:
  ```bash
  make test
  ```

---

## Bootstrapping Inicial (Crear Workspace y Proyecto)

Dado que las memorias requieren estar enlazadas a un Workspace y Proyecto existentes, puedes usar el CLI para inicializarlos antes de guardar registros:

1. **Crear Workspace**:
   ```bash
   ./bin/memory workspace create --slug "default" --name "Workspace por Defecto"
   ```
2. **Crear Proyecto**:
   ```bash
   ./bin/memory project create --workspace "default" --slug "my-app"
   ```

---

## Uso del CLI (`memory`)

Una vez compilado el CLI con `make build` y asignado en tu `PATH` o ejecutándolo como `./bin/memory`:

- **Guardar una memoria en lenguaje natural**:
  ```bash
  ./bin/memory remember "Configuramos el pool de conexiones en Go usando pgxpool para mejorar rendimiento de concurrencia." --workspace "default" --project "my-app" --tags go,db,performance --source developer
  ```
- **Búsqueda semántica (basada en similitud del coseno)**:
  ```bash
  ./bin/memory search "pool de conexiones en Go" --workspace "default" --project "my-app" --limit 3
  ```
- **Obtener contexto listo para un LLM**:
  ```bash
  ./bin/memory context "go connection pool configuration" --workspace "default" --project "my-app" --limit 5
  ```

---

## Buenas Prácticas y Decisiones de Diseño Adicionales

- **Las Interfaces Pertenecen al Consumidor**: Las interfaces de acceso a datos se definen en `internal/service/` en lugar de `internal/domain/`. Esto previene el acoplamiento innecesario y permite testear las clases simulando repositorios ficticios.
- **Errores del Dominio**: El backend utiliza errores explícitos (ej. `ErrMemoryNotFound`) definidos en el paquete `domain`. Las capas de persistencia mapean los errores específicos del motor de base de datos a estos tipos puros, y la capa HTTP los traduce a códigos semánticos.
- **Seguridad y DoS**: Las peticiones entrantes están limitadas a un tamaño máximo de 5MB mediante el middleware `MaxBytesMiddleware` para mitigar ataques por consumo excesivo de memoria en payloads JSON.
- **SQL Sin ORM**: Para garantizar el rendimiento del ordenamiento vectorial tridimensional (`pgvector`) y evitar dependencias complejas, se optó por escribir consultas SQL puras y parametrizadas directamente sobre `pgxpool`.
- **Uso Obligatorio de `context.Context`**: Toda llamada externa, consulta a la base de datos o generación de embeddings recibe y transfiere el contexto para garantizar el soporte de cancelaciones, timeouts y trazas a lo largo del ciclo de vida del request.
