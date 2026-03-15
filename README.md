# SuperMercadoUP

Sistema cliente-servidor concurrente en **Go** que simula un motor de inventario y procesamiento de órdenes, inspirado en una plataforma de e-commerce.

---

## Descripción

**SuperMercadoUP** es un sistema basado en **TCP** que permite administrar productos, gestionar carritos de compra y procesar órdenes de manera concurrente.

El proyecto está dividido en tres componentes principales:

- **Servidor**: centraliza la lógica de negocio, administra el inventario compartido y registra eventos en `server.log`
- **Cliente**: permite consultar productos, agregarlos al carrito y generar órdenes
- **Administrador**: permite agregar productos, actualizar stock y precios, y consultar el historial de órdenes

El sistema fue desarrollado para trabajar con múltiples conexiones simultáneas y mantener consistencia en el inventario compartido.

---

## Características principales

- Comunicación cliente-servidor mediante **TCP**
- Soporte para **múltiples clientes concurrentes**
- Gestión de inventario compartido
- Carrito de compras por cliente
- Creación y consulta de órdenes
- Panel de administrador para gestión del inventario
- Registro de eventos en archivo log
- **Bloqueo de productos en carrito**
- **Persistencia del estado del sistema**

---

## Estructura del proyecto

```bash
SuperMercadoUP/
├── cmd/
│   ├── admin/
│   │   └── admin.go
│   ├── client/
│   │   └── client.go
│   └── server/
│       └── server.go
├── internal/
│   ├── logger/
│   │   └── logger.go
│   └── models/
│       └── models.go
├── go.mod
└── README.md
```

---

## Requisitos

- **Go 1.21** o superior

---

## Instalación

Clona el repositorio:

```bash
git clone https://github.com/DemianVGLl2/SuperMercadoUP.git
cd SuperMercadoUP
```

---

## Ejecución

### 1. Iniciar el servidor

En una terminal:

```bash
go run cmd/server/server.go
```

El servidor escuchará en:

```
localhost:8000
```

Además, generará el archivo:

```
server.log
```

---

### 2. Iniciar un cliente

En otra terminal:

```bash
go run cmd/client/client.go
```

---

### 3. Iniciar el administrador

En una tercera terminal:

```bash
go run cmd/admin/admin.go
```

---

## Uso del sistema

Se pueden conectar múltiples clientes y administradores al mismo tiempo.

### Funcionalidades del cliente

- Listar productos
- Agregar productos al carrito
- Ver carrito
- Realizar orden
- Consultar ayuda
- Salir

### Funcionalidades del administrador

- Agregar productos
- Actualizar stock
- Actualizar precio
- Ver historial de órdenes
- Listar productos
- Salir

---

## Ejemplos de uso

### Cliente: listar productos

```
Choose an option: 1
Available products:
ID | Name | Price | Stock
--------------------------
1 | Leche | 28.50 | 10
2 | Pan | 15.00 | 8
3 | Huevos | 42.00 | 12
```

### Cliente: agregar al carrito

```
Choose an option: 2
Enter product ID: 1
Enter quantity: 2
OK producto agregado al carrito
```

### Cliente: ver carrito

```
Choose an option: 3
Cart contents:
ProductID | Name | Qty | UnitPrice | Subtotal
---------------------------------------------
1 | Leche | 2 | 28.50 | 57.00
Total: 57.00
```

### Cliente: realizar orden

```
Choose an option: 4
OK
Orden creada con id 1
Total: 57.00
Status: COMPLETED
```

### Administrador: agregar producto

```
Choose an option: 1
Enter product ID: 4
Enter product name: Mantequilla
Enter price: 35.0
Enter stock: 20
OK producto agregado
```

### Administrador: actualizar stock

```
Choose an option: 2
Enter product ID: 1
Enter new stock: 50
OK stock actualizado
```

### Administrador: actualizar precio

```
Choose an option: 3
Enter product ID: 2
Enter new price: 18.5
OK precio actualizado
```

### Administrador: ver historial de órdenes

```
Choose an option: 4
Order history:
--------------
OrderID:1 Total:57.00 Status:COMPLETED
OrderID:2 Total:84.00 Status:COMPLETED
```

---

## Concurrencia y consistencia

El sistema utiliza mecanismos de sincronización para proteger el acceso concurrente al inventario y a las órdenes.

Además, se implementó un esquema de **bloqueo de productos en carrito**, de forma que cuando un cliente agrega un producto, ese stock queda reservado temporalmente y no puede ser tomado por otro cliente hasta que:

- se complete la orden
- se libere el carrito
- el cliente se desconecte

Esto ayuda a mantener consistencia bajo concurrencia.

---

## Mejoras implementadas

### Bloqueo de productos en carrito

Cuando un cliente agrega un producto a su carrito, la cantidad seleccionada queda bloqueada temporalmente para evitar conflictos con otros clientes que intenten comprar el mismo producto al mismo tiempo.

### Persistencia del estado

El sistema permite conservar el estado del inventario y de las órdenes para evitar pérdida de datos al reiniciar el servidor.

---

## Trabajo futuro

- Registro e inicio de sesión de usuarios
- Historial de compras por usuario
- Control de acceso por roles
- Cancelación de órdenes
- Eliminación de productos del carrito
- Liberación automática de productos bloqueados mediante timeout
- Interfaz gráfica para cliente y administrador

---

## Tecnologías usadas

- **Go**
- **TCP sockets**
- **Goroutines**
- **Mutex / sincronización concurrente**
- **Arquitectura cliente-servidor**

---

## Autores

Proyecto desarrollado por el equipo **SuperMercadoUP** para la materia **Advanced Parallel Programming**.
