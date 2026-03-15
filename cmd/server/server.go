package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"unicode"

	"github.com/DemianVGLl2/SuperMercadoUP/internal/logger"
	"github.com/DemianVGLl2/SuperMercadoUP/internal/models"
)

func main() {
	if err := logger.Init("server.log"); err != nil {
		log.Fatal("Error iniciando logger: ", err)
	}

	store := models.NewStore()

	// productos de prueba
	store.Products[1] = &models.Product{ID: 1, Name: "Leche", Price: 28.5, Stock: 10, Blocked: 0}
	store.Products[2] = &models.Product{ID: 2, Name: "Pan", Price: 15.0, Stock: 8, Blocked: 0}
	store.Products[3] = &models.Product{ID: 3, Name: "Huevos", Price: 42.0, Stock: 12, Blocked: 0}

	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Listening to: " + listener.Addr().String())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(conn, store)
	}
}

func handleConn(conn net.Conn, store *models.Store) {
	defer conn.Close()

	addr := conn.RemoteAddr().String()
	logger.Log("SERVER", "CONNECT", addr)

	cart := models.Cart{}
	input := bufio.NewScanner(conn)

	for input.Scan() {
		line := strings.TrimSpace(input.Text())
		if line == "" {
			continue
		}

		trozos := strings.Fields(line)
		cmd := strings.ToUpper(trozos[0])

		switch cmd {
		case "LIST_PRODUCTS":
			sacarProductos(conn, store)

		case "ADD_TO_CART":
			agregarAlCarrito(conn, store, &cart, trozos)

		case "VIEW_CART":
			verCarrito(conn, store, &cart)

		case "PLACE_ORDER":
			hacerOrden(conn, store, &cart)

		case "ADD_PRODUCT":
			agregarProducto(conn, store, trozos)

		case "UPDATE_STOCK":
			actualizarStock(conn, store, trozos)

		case "UPDATE_PRICE":
			actualizarPrecio(conn, store, trozos)

		case "LIST_ORDERS":
			listarOrdenes(conn, store)

		case "HELP":
			fmt.Fprintln(conn, "OK")
			fmt.Fprintln(conn, "LIST_PRODUCTS")
			fmt.Fprintln(conn, "ADD_TO_CART <productID> <quantity>")
			fmt.Fprintln(conn, "VIEW_CART")
			fmt.Fprintln(conn, "PLACE_ORDER")
			fmt.Fprintln(conn, "ADD_PRODUCT <id> <name> <price> <stock>")
			fmt.Fprintln(conn, "UPDATE_STOCK <id> <newStock>")
			fmt.Fprintln(conn, "UPDATE_PRICE <id> <newPrice>")
			fmt.Fprintln(conn, "LIST_ORDERS")
			fmt.Fprintln(conn, "HELP")
			fmt.Fprintln(conn, "EXIT")
			fmt.Fprintln(conn, "END")

		case "EXIT":
			liberarCarrito(store, &cart)
			fmt.Fprintln(conn, "OK Bye")
			logger.Log("SERVER", "DISCONNECT", addr)
			return

		default:
			fmt.Fprintln(conn, "ERROR comando no valido")
			fmt.Fprintln(conn, "END")
		}
	}

	if err := input.Err(); err != nil {
		log.Println("Read error:", err)
	}

	// Si el cliente se desconecta sin comprar, liberar bloqueos
	liberarCarrito(store, &cart)
	logger.Log("SERVER", "DISCONNECT", addr)
}

func sacarProductos(conn net.Conn, store *models.Store) {
	store.Mu.RLock()
	defer store.Mu.RUnlock()

	fmt.Fprintln(conn, "OK")

	if len(store.Products) == 0 {
		fmt.Fprintln(conn, "END")
		return
	}

	for _, p := range store.Products {
		disponible := p.Stock - p.Blocked
		if disponible < 0 {
			disponible = 0
		}

		// Mostramos el stock disponible real, no el total físico
		fmt.Fprintf(conn, "%d|%s|%.2f|%d\n", p.ID, p.Name, p.Price, disponible)
	}

	fmt.Fprintln(conn, "END")
}

func agregarAlCarrito(conn net.Conn, store *models.Store, cart *models.Cart, trozos []string) {
	if len(trozos) != 3 {
		fmt.Fprintln(conn, "ERROR usa: ADD_TO_CART <productID> <quantity>")
		fmt.Fprintln(conn, "END")
		return
	}

	pid, err1 := strconv.Atoi(trozos[1])
	cant, err2 := strconv.Atoi(trozos[2])

	if err1 != nil || err2 != nil {
		fmt.Fprintln(conn, "ERROR datos invalidos")
		fmt.Fprintln(conn, "END")
		return
	}

	if cant <= 0 {
		fmt.Fprintln(conn, "ERROR la cantidad debe ser mayor a 0")
		fmt.Fprintln(conn, "END")
		return
	}

	store.Mu.Lock()
	defer store.Mu.Unlock()

	prod, ok := store.Products[pid]
	if !ok {
		fmt.Fprintln(conn, "ERROR ese producto no existe")
		fmt.Fprintln(conn, "END")
		return
	}

	disponible := prod.Stock - prod.Blocked
	if cant > disponible {
		fmt.Fprintln(conn, "ERROR producto bloqueado o sin stock disponible")
		fmt.Fprintln(conn, "END")
		return
	}

	for i := range cart.Items {
		if cart.Items[i].ProductID == pid {
			cart.Items[i].Quantity += cant
			prod.Blocked += cant

			logger.Log("CLIENT", "ADD_TO_CART", fmt.Sprintf("product=%d qty=%d blocked=%d", pid, cant, prod.Blocked))
			fmt.Fprintln(conn, "OK producto agregado al carrito y bloqueado")
			fmt.Fprintln(conn, "END")
			return
		}
	}

	cart.Items = append(cart.Items, models.OrderItem{
		ProductID: pid,
		Quantity:  cant,
		UnitPrice: prod.Price,
	})

	prod.Blocked += cant

	logger.Log("CLIENT", "ADD_TO_CART", fmt.Sprintf("product=%d qty=%d blocked=%d", pid, cant, prod.Blocked))
	fmt.Fprintln(conn, "OK producto agregado al carrito y bloqueado")
	fmt.Fprintln(conn, "END")
}

func verCarrito(conn net.Conn, store *models.Store, cart *models.Cart) {
	store.Mu.RLock()
	defer store.Mu.RUnlock()

	fmt.Fprintln(conn, "OK")

	if len(cart.Items) == 0 {
		fmt.Fprintln(conn, "END")
		return
	}

	for _, item := range cart.Items {
		nombre := "Producto"
		if prod, ok := store.Products[item.ProductID]; ok {
			nombre = prod.Name
		}

		fmt.Fprintf(conn, "%d|%s|%d|%.2f\n",
			item.ProductID,
			nombre,
			item.Quantity,
			item.UnitPrice,
		)
	}

	fmt.Fprintln(conn, "END")
}

func hacerOrden(conn net.Conn, store *models.Store, cart *models.Cart) {
	store.Mu.Lock()
	defer store.Mu.Unlock()

	if len(cart.Items) == 0 {
		fmt.Fprintln(conn, "ERROR el carrito esta vacio")
		fmt.Fprintln(conn, "END")
		return
	}

	total := 0.0

	for _, item := range cart.Items {
		prod, ok := store.Products[item.ProductID]
		if !ok {
			fmt.Fprintf(conn, "ERROR el producto %d ya no existe\n", item.ProductID)
			fmt.Fprintln(conn, "END")
			return
		}

		// Como ya está bloqueado para este carrito, aquí solo validamos stock físico
		if item.Quantity > prod.Stock {
			fmt.Fprintf(conn, "ERROR no hay stock suficiente para %s\n", prod.Name)
			fmt.Fprintln(conn, "END")
			return
		}

		total += float64(item.Quantity) * item.UnitPrice
	}

	for _, item := range cart.Items {
		prod := store.Products[item.ProductID]
		prod.Stock -= item.Quantity
		prod.Blocked -= item.Quantity

		if prod.Blocked < 0 {
			prod.Blocked = 0
		}
	}

	store.NextOrderID++
	idOrden := store.NextOrderID

	itemsCopia := make([]models.OrderItem, len(cart.Items))
	copy(itemsCopia, cart.Items)

	orden := &models.Order{
		ID:     idOrden,
		Items:  itemsCopia,
		Total:  total,
		Status: models.Completed,
	}

	store.Orders[idOrden] = orden
	cart.Items = nil

	logger.Log("CLIENT", "PLACE_ORDER", fmt.Sprintf("order=%d total=%.2f", orden.ID, orden.Total))

	fmt.Fprintln(conn, "OK")
	fmt.Fprintf(conn, "Orden creada con id %d\n", orden.ID)
	fmt.Fprintf(conn, "Total: %.2f\n", orden.Total)
	fmt.Fprintf(conn, "Status: %s\n", orden.Status)
	fmt.Fprintln(conn, "END")
}

func liberarCarrito(store *models.Store, cart *models.Cart) {
	store.Mu.Lock()
	defer store.Mu.Unlock()

	if len(cart.Items) == 0 {
		return
	}

	for _, item := range cart.Items {
		prod, ok := store.Products[item.ProductID]
		if !ok {
			continue
		}

		prod.Blocked -= item.Quantity
		if prod.Blocked < 0 {
			prod.Blocked = 0
		}
	}

	cart.Items = nil
}

func nombreValido(nombre string) bool {
	if strings.TrimSpace(nombre) == "" {
		return false
	}

	for _, r := range nombre {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			continue
		}
		return false
	}

	return true
}

func agregarProducto(conn net.Conn, store *models.Store, trozos []string) {
	store.Mu.Lock()
	defer store.Mu.Unlock()

	if len(trozos) < 5 {
		fmt.Fprintln(conn, "ERROR usa: ADD_PRODUCT <id> <name> <price> <stock>")
		fmt.Fprintln(conn, "END")
		return
	}

	id, err1 := strconv.Atoi(trozos[1])
	precio, err2 := strconv.ParseFloat(trozos[len(trozos)-2], 64)
	stock, err3 := strconv.Atoi(trozos[len(trozos)-1])
	nombre := strings.Join(trozos[2:len(trozos)-2], " ")

	if err1 != nil || err2 != nil || err3 != nil {
		fmt.Fprintln(conn, "ERROR datos invalidos")
		fmt.Fprintln(conn, "END")
		return
	}

	if id <= 0 {
		fmt.Fprintln(conn, "ERROR id invalido")
		fmt.Fprintln(conn, "END")
		return
	}

	if !nombreValido(nombre) {
		fmt.Fprintln(conn, "ERROR nombre invalido")
		fmt.Fprintln(conn, "END")
		return
	}

	if precio < 0 {
		fmt.Fprintln(conn, "ERROR el precio no puede ser negativo")
		fmt.Fprintln(conn, "END")
		return
	}

	if stock < 0 {
		fmt.Fprintln(conn, "ERROR el stock no puede ser negativo")
		fmt.Fprintln(conn, "END")
		return
	}

	if _, existe := store.Products[id]; existe {
		fmt.Fprintln(conn, "ERROR ya existe un producto con ese ID")
		fmt.Fprintln(conn, "END")
		return
	}

	store.Products[id] = &models.Product{
		ID:      id,
		Name:    nombre,
		Price:   precio,
		Stock:   stock,
		Blocked: 0,
	}

	logger.Log("ADMIN", "ADD_PRODUCT", fmt.Sprintf("id=%d name=%s price=%.2f stock=%d", id, nombre, precio, stock))

	fmt.Fprintln(conn, "OK producto agregado")
	fmt.Fprintln(conn, "END")
}

func actualizarStock(conn net.Conn, store *models.Store, trozos []string) {
	store.Mu.Lock()
	defer store.Mu.Unlock()

	if len(trozos) != 3 {
		fmt.Fprintln(conn, "ERROR usa: UPDATE_STOCK <id> <newStock>")
		fmt.Fprintln(conn, "END")
		return
	}

	id, err1 := strconv.Atoi(trozos[1])
	nuevoStock, err2 := strconv.Atoi(trozos[2])

	if err1 != nil || err2 != nil {
		fmt.Fprintln(conn, "ERROR datos invalidos")
		fmt.Fprintln(conn, "END")
		return
	}

	if nuevoStock < 0 {
		fmt.Fprintln(conn, "ERROR el stock no puede ser negativo")
		fmt.Fprintln(conn, "END")
		return
	}

	prod, ok := store.Products[id]
	if !ok {
		fmt.Fprintln(conn, "ERROR el producto no existe")
		fmt.Fprintln(conn, "END")
		return
	}

	if nuevoStock < prod.Blocked {
		fmt.Fprintln(conn, "ERROR no puedes poner stock menor que la cantidad bloqueada")
		fmt.Fprintln(conn, "END")
		return
	}

	prod.Stock = nuevoStock

	logger.Log("ADMIN", "UPDATE_STOCK", fmt.Sprintf("id=%d newStock=%d", id, nuevoStock))

	fmt.Fprintln(conn, "OK stock actualizado")
	fmt.Fprintln(conn, "END")
}

func actualizarPrecio(conn net.Conn, store *models.Store, trozos []string) {
	store.Mu.Lock()
	defer store.Mu.Unlock()

	if len(trozos) != 3 {
		fmt.Fprintln(conn, "ERROR usa: UPDATE_PRICE <id> <newPrice>")
		fmt.Fprintln(conn, "END")
		return
	}

	id, err1 := strconv.Atoi(trozos[1])
	nuevoPrecio, err2 := strconv.ParseFloat(trozos[2], 64)

	if err1 != nil || err2 != nil {
		fmt.Fprintln(conn, "ERROR datos invalidos")
		fmt.Fprintln(conn, "END")
		return
	}

	if nuevoPrecio < 0 {
		fmt.Fprintln(conn, "ERROR el precio no puede ser negativo")
		fmt.Fprintln(conn, "END")
		return
	}

	prod, ok := store.Products[id]
	if !ok {
		fmt.Fprintln(conn, "ERROR el producto no existe")
		fmt.Fprintln(conn, "END")
		return
	}

	prod.Price = nuevoPrecio

	logger.Log("ADMIN", "UPDATE_PRICE", fmt.Sprintf("id=%d newPrice=%.2f", id, nuevoPrecio))

	fmt.Fprintln(conn, "OK precio actualizado")
	fmt.Fprintln(conn, "END")
}

func listarOrdenes(conn net.Conn, store *models.Store) {
	store.Mu.RLock()
	defer store.Mu.RUnlock()

	fmt.Fprintln(conn, "OK")

	if len(store.Orders) == 0 {
		fmt.Fprintln(conn, "END")
		return
	}

	for _, orden := range store.Orders {
		fmt.Fprintf(conn, "OrderID:%d Total:%.2f Status:%s\n", orden.ID, orden.Total, orden.Status)
	}

	logger.Log("ADMIN", "LIST_ORDERS", "consulta de historial de ordenes")

	fmt.Fprintln(conn, "END")
}
