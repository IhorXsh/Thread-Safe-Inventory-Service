package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/IhorXsh/Thread-Safe-Inventory-Service/internal/inventory"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	svc := inventory.NewSafeInventoryService(map[inventory.ProductID]*inventory.Product{
		"A": inventory.NewProduct("A", "Widget A", 10),
		"B": inventory.NewProduct("B", "Widget B", 5),
	})

	switch os.Args[1] {
	case "get":
		if len(os.Args) != 3 {
			usage()
			os.Exit(2)
		}
		id := os.Args[2]
		fmt.Printf("%s stock: %d\n", id, svc.GetStock(id))
	case "reserve":
		if len(os.Args) != 4 {
			usage()
			os.Exit(2)
		}
		id := os.Args[2]
		qty, err := strconv.ParseUint(os.Args[3], 10, 64)
		if err != nil || qty == 0 {
			fmt.Fprintln(os.Stderr, "quantity must be a positive integer")
			os.Exit(2)
		}
		if err := svc.Reserve(id, qty); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("reserved %d of %s\n", qty, id)
	case "reserve-multiple":
		if len(os.Args) != 3 {
			usage()
			os.Exit(2)
		}
		items, err := parseItems(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		if err := svc.ReserveMultiple(items); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("reserved all items")
	default:
		usage()
		os.Exit(2)
	}
}

func parseItems(arg string) ([]inventory.ReserveItem, error) {
	parts := strings.Split(arg, ",")
	items := make([]inventory.ReserveItem, 0, len(parts))

	for _, part := range parts {
		p := strings.Split(part, ":")
		if len(p) != 2 {
			return nil, fmt.Errorf("invalid item %q, expected ID:QTY", part)
		}
		qty, err := strconv.ParseUint(p[1], 10, 64)
		if err != nil || qty == 0 {
			return nil, fmt.Errorf("invalid quantity for %q", part)
		}
		items = append(items, inventory.ReserveItem{
			ProductID: inventory.ProductID(p[0]),
			Quantity:  qty,
		})
	}

	return items, nil
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  inventory-service get <product-id>")
	fmt.Println("  inventory-service reserve <product-id> <qty>")
	fmt.Println("  inventory-service reserve-multiple <id:qty,id:qty,...>")
}
