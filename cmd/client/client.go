package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const serverAddress = "localhost:8000"

type Product struct {
	ID    int
	Name  string
	Price float64
	Stock int
}

type CartItem struct {
	ProductID int
	Name      string
	Quantity  int
	UnitPrice float64
}

func main() {
	fmt.Println("===================================")
	fmt.Println("      SuperMercadoUP - Client")
	fmt.Println("===================================")

	conn, err := net.DialTimeout("tcp", serverAddress, 5*time.Second)
	if err != nil {
		fmt.Println("Error: could not connect to server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to server:", serverAddress)

	reader := bufio.NewReader(conn)
	console := bufio.NewReader(os.Stdin)

	for {
		showMenu()

		fmt.Print("Choose an option: ")
		option, err := console.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}
		option = strings.TrimSpace(option)

		switch option {
		case "1":
			listProducts(conn, reader)
		case "2":
			addToCart(conn, reader, console)
		case "3":
			viewCart(conn, reader)
		case "4":
			placeOrder(conn, reader)
		case "5":
			showHelp(conn, reader)
		case "0":
			exitClient(conn, reader)
			fmt.Println("Disconnected from server.")
			return
		default:
			fmt.Println("Invalid option. Try again.")
		}
	}
}

func showMenu() {
	fmt.Println()
	fmt.Println("------------- MENU -------------")
	fmt.Println("1) List products")
	fmt.Println("2) Add product to cart")
	fmt.Println("3) View cart")
	fmt.Println("4) Place order")
	fmt.Println("5) Help")
	fmt.Println("0) Exit")
	fmt.Println("--------------------------------")
}

func listProducts(conn net.Conn, reader *bufio.Reader) {
	err := sendCommand(conn, "LIST_PRODUCTS")
	if err != nil {
		fmt.Println("Error sending command:", err)
		return
	}

	lines, err := readMultiLineResponse(reader)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	if len(lines) == 0 {
		fmt.Println("Empty response from server.")
		return
	}

	if strings.HasPrefix(lines[0], "ERROR") {
		fmt.Println(lines[0])
		return
	}

	if lines[0] != "OK" {
		fmt.Println("Unexpected response:", lines[0])
		return
	}

	if len(lines) == 1 {
		fmt.Println("No products found.")
		return
	}

	fmt.Println()
	fmt.Println("Available products:")
	fmt.Println("ID | Name | Price | Stock")
	fmt.Println("--------------------------")

	for _, line := range lines[1:] {
		product, err := parseProduct(line)
		if err != nil {
			fmt.Println("Skipping invalid product line:", line)
			continue
		}
		fmt.Printf("%d | %s | %.2f | %d\n", product.ID, product.Name, product.Price, product.Stock)
	}
}

func addToCart(conn net.Conn, reader *bufio.Reader, console *bufio.Reader) {
	productID, ok := askInt(console, "Enter product ID: ")
	if !ok {
		return
	}

	quantity, ok := askInt(console, "Enter quantity: ")
	if !ok {
		return
	}

	if quantity <= 0 {
		fmt.Println("Quantity must be greater than 0.")
		return
	}

	command := fmt.Sprintf("ADD_TO_CART %d %d", productID, quantity)
	err := sendCommand(conn, command)
	if err != nil {
		fmt.Println("Error sending command:", err)
		return
	}

	lines, err := readMultiLineResponse(reader)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	for _, line := range lines {
		fmt.Println(line)
	}
}

func viewCart(conn net.Conn, reader *bufio.Reader) {
	err := sendCommand(conn, "VIEW_CART")
	if err != nil {
		fmt.Println("Error sending command:", err)
		return
	}

	lines, err := readMultiLineResponse(reader)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	if len(lines) == 0 {
		fmt.Println("Empty response from server.")
		return
	}

	if strings.HasPrefix(lines[0], "ERROR") {
		fmt.Println(lines[0])
		return
	}

	if lines[0] != "OK" {
		fmt.Println("Unexpected response:", lines[0])
		return
	}

	if len(lines) == 1 {
		fmt.Println("Cart is empty.")
		return
	}

	fmt.Println()
	fmt.Println("Cart contents:")
	fmt.Println("ProductID | Name | Qty | UnitPrice | Subtotal")
	fmt.Println("---------------------------------------------")

	total := 0.0

	for _, line := range lines[1:] {
		item, err := parseCartItem(line)
		if err != nil {
			fmt.Println("Skipping invalid cart line:", line)
			continue
		}
		subtotal := float64(item.Quantity) * item.UnitPrice
		total += subtotal
		fmt.Printf("%d | %s | %d | %.2f | %.2f\n",
			item.ProductID, item.Name, item.Quantity, item.UnitPrice, subtotal)
	}

	fmt.Printf("Total: %.2f\n", total)
}

func placeOrder(conn net.Conn, reader *bufio.Reader) {
	err := sendCommand(conn, "PLACE_ORDER")
	if err != nil {
		fmt.Println("Error sending command:", err)
		return
	}

	lines, err := readMultiLineResponse(reader)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	if len(lines) == 0 {
		fmt.Println("Empty response from server.")
		return
	}

	for _, line := range lines {
		fmt.Println(line)
	}
}

func showHelp(conn net.Conn, reader *bufio.Reader) {
	err := sendCommand(conn, "HELP")
	if err != nil {
		fmt.Println("Error sending command:", err)
		return
	}

	lines, err := readMultiLineResponse(reader)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	if len(lines) == 0 {
		fmt.Println("No help received from server.")
		return
	}

	fmt.Println()
	for _, line := range lines {
		fmt.Println(line)
	}
}

func exitClient(conn net.Conn, reader *bufio.Reader) {
	_ = sendCommand(conn, "EXIT")
	_, _ = readSingleLineResponse(reader)
}

func sendCommand(conn net.Conn, command string) error {
	_, err := fmt.Fprintf(conn, "%s\n", command)
	return err
}

func readSingleLineResponse(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func readMultiLineResponse(reader *bufio.Reader) ([]string, error) {
	var lines []string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)

		if line == "END" {
			break
		}

		lines = append(lines, line)
	}

	return lines, nil
}

func askInt(console *bufio.Reader, prompt string) (int, bool) {
	fmt.Print(prompt)
	text, err := console.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		return 0, false
	}

	text = strings.TrimSpace(text)
	value, err := strconv.Atoi(text)
	if err != nil {
		fmt.Println("Please enter a valid integer.")
		return 0, false
	}

	return value, true
}

func parseProduct(line string) (Product, error) {
	parts := strings.Split(line, "|")
	if len(parts) != 4 {
		return Product{}, fmt.Errorf("invalid product format")
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		return Product{}, err
	}

	price, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return Product{}, err
	}

	stock, err := strconv.Atoi(parts[3])
	if err != nil {
		return Product{}, err
	}

	return Product{
		ID:    id,
		Name:  parts[1],
		Price: price,
		Stock: stock,
	}, nil
}

func parseCartItem(line string) (CartItem, error) {
	parts := strings.Split(line, "|")
	if len(parts) != 4 {
		return CartItem{}, fmt.Errorf("invalid cart format")
	}

	productID, err := strconv.Atoi(parts[0])
	if err != nil {
		return CartItem{}, err
	}

	quantity, err := strconv.Atoi(parts[2])
	if err != nil {
		return CartItem{}, err
	}

	unitPrice, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return CartItem{}, err
	}

	return CartItem{
		ProductID: productID,
		Name:      parts[1],
		Quantity:  quantity,
		UnitPrice: unitPrice,
	}, nil
}
