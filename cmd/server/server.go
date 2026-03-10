package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/DemianVGLl2/SuperMercadoUP/internal/logger"
	"github.com/DemianVGLl2/SuperMercadoUP/internal/models"
)

func main() {
	if err := logger.Init("server.log"); err != nil {
		log.Fatal("Error iniciando logger: ", err)
	}

	store := models.NewStore()

	store.Products[1] = &models.Product{ID: 1, Name: "Leche", Price: 28.5, Stock: 10}
	store.Products[2] = &models.Product{ID: 2, Name: "Pan", Price: 15.0, Stock: 8}
	store.Products[3] = &models.Product{ID: 3, Name: "Huevos", Price: 42.0, Stock: 12}

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

		case "HELP":
			fmt.Fprintln(conn, "OK")
			fmt.Fprintln(conn, "LIST_PRODUCTS")
			fmt.Fprintln(conn, "ADD_TO_CART <productID> <quantity>")
			fmt.Fprintln(conn, "VIEW_CART")
			fmt.Fprintln(conn, "PLACE_ORDER")
			fmt.Fprintln(conn, "HELP")
			fmt.Fprintln(conn, "EXIT")
			fmt.Fprintln(conn, "END")

		case "EXIT":
			fmt.Fprintln(conn, "OK Bye")
			logger.Log("SERVER", "DISCONNECT", addr)
			return

		default:
			fmt.Fprintln(conn, "ERROR comando no valido")
		}
	}

	if err := input.Err(); err != nil {
		log.Println("Read error:", err)
	}

	logger.Log("SERVER", "DISCONNECT", addr)
}

func sacarProductos(conn net.Conn, store *models.Store) {
	fmt.Fprintln(conn, "OK")

	if len(store.Products) == 0 {
		fmt.Fprintln(conn, "END")
		return
	}

	for _, p := range store.Products {
		fmt.Fprintf(conn, "%d|%s|%.2f|%d\n", p.ID, p.Name, p.Price, p.Stock)
	}

	fmt.Fprintln(conn, "END")
}

func agregarAlCarrito(conn net.Conn, store *models.Store, cart *models.Cart, trozos []string) {
	if len(trozos) != 3 {
		fmt.Fprintln(conn, "ERROR usa: ADD_TO_CART <productID> <quantity>")
		return
	}

	pid, err1 := strconv.Atoi(trozos[1])
	cant, err2 := strconv.Atoi(trozos[2])

	if err1 != nil || err2 != nil {
		fmt.Fprintln(conn, "ERROR datos invalidos")
		return
	}

	if cant <= 0 {
		fmt.Fprintln(conn, "ERROR la cantidad debe ser mayor a 0")
		return
	}

	prod, ok := store.Products[pid]
	if !ok {
		fmt.Fprintln(conn, "ERROR ese producto no existe")
		return
	}

	if cant > prod.Stock {
		fmt.Fprintln(conn, "ERROR no hay suficiente stock")
		return
	}

	for i := range cart.Items {
		if cart.Items[i].ProductID == pid {
			nuevaCant := cart.Items[i].Quantity + cant

			if nuevaCant > prod.Stock {
				fmt.Fprintln(conn, "ERROR no alcanza el stock para agregar mas")
				return
			}

			cart.Items[i].Quantity = nuevaCant
			fmt.Fprintln(conn, "OK producto agregado al carrito")
			return
		}
	}

	cart.Items = append(cart.Items, models.OrderItem{
		ProductID: pid,
		Quantity:  cant,
		UnitPrice: prod.Price,
	})

	fmt.Fprintln(conn, "OK producto agregado al carrito")
}

func verCarrito(conn net.Conn, store *models.Store, cart *models.Cart) {
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

		if item.Quantity > prod.Stock {
			fmt.Fprintf(conn, "ERROR no hay stock suficiente para %s\n", prod.Name)
			fmt.Fprintln(conn, "END")
			return
		}

		total += float64(item.Quantity) * item.UnitPrice
	}

	// ya que validó, ahora descuenta stock
	for _, item := range cart.Items {
		store.Products[item.ProductID].Stock -= item.Quantity
	}

	idOrden := len(store.Orders) + 1

	orden := &models.Order{
		ID:     idOrden,
		Items:  cart.Items,
		Total:  total,
		Status: models.Completed,
	}

	store.Orders[idOrden] = orden

	cart.Items = nil

	fmt.Fprintln(conn, "OK")
	fmt.Fprintf(conn, "Orden creada con id %d\n", orden.ID)
	fmt.Fprintf(conn, "Total: %.2f\n", orden.Total)
	fmt.Fprintf(conn, "Status: %s\n", orden.Status)
	fmt.Fprintln(conn, "END")
}
