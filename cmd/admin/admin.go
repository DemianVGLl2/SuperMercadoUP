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

func main() {
	fmt.Println("===================================")
	fmt.Println("      SuperMercadoUP - Admin")
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
			addProduct(conn, reader, console)
		case "2":
			updateStock(conn, reader, console)
		case "3":
			updatePrice(conn, reader, console)
		case "4":
			listOrders(conn, reader)
		case "5":
			listProducts(conn, reader)
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
	fmt.Println("1) Add products")
	fmt.Println("2) Update stock")
	fmt.Println("3) Update price")
	fmt.Println("4) List orders")
	fmt.Println("5) List products")
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

func addProduct(conn net.Conn, reader *bufio.Reader, console *bufio.Reader) {
	id, ok := askInt(console, "Enter product ID: ")
	if !ok {
		return
	}

	fmt.Print("Enter product name: ")
	name, _ := console.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Enter price: ")
	priceStr, _ := console.ReadString('\n')
	priceStr = strings.TrimSpace(priceStr)

	stock, ok := askInt(console, "Enter stock: ")
	if !ok {
		return
	}

	command := fmt.Sprintf("ADD_PRODUCT %d %s %s %d", id, name, priceStr, stock)
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

func updateStock(conn net.Conn, reader *bufio.Reader, console *bufio.Reader) {
	id, ok := askInt(console, "Enter product ID: ")
	if !ok {
		return
	}

	newStock, ok := askInt(console, "Enter new stock: ")
	if !ok {
		return
	}

	command := fmt.Sprintf("UPDATE_STOCK %d %d", id, newStock)
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

func updatePrice(conn net.Conn, reader *bufio.Reader, console *bufio.Reader) {
	id, ok := askInt(console, "Enter product ID: ")
	if !ok {
		return
	}

	fmt.Print("Enter new price: ")
	priceStr, _ := console.ReadString('\n')
	priceStr = strings.TrimSpace(priceStr)

	command := fmt.Sprintf("UPDATE_PRICE %d %s", id, priceStr)
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

func listOrders(conn net.Conn, reader *bufio.Reader) {
	err := sendCommand(conn, "LIST_ORDERS")
	if err != nil {
		fmt.Println("Error sending command:", err)
		return
	}

	lines, err := readMultiLineResponse(reader)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "OK") {
		fmt.Println("No orders found.")
		return
	}

	fmt.Println()
	fmt.Println("Order history:")
	fmt.Println("--------------")
	for _, line := range lines[1:] {
		fmt.Println(line)
	}
}
