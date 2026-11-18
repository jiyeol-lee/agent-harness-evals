// Command shop prices orders and reserves their stock from a JSON stock file.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"example.com/shop/events"
	"example.com/shop/orders"
	"example.com/shop/pricing"
	"example.com/shop/stock"
)

const usage = `usage: shop [-f file] [-id ID] <command>

commands:
  init                      write a fresh stock file with 100 apples and 50 pears
  count NAME                print how many NAME are in stock
  price UNIT QTY            print the price of QTY units at UNIT each
  order NAME=QTY@PRICE ...  reserve stock for every line and print the total

flags:
`

func main() {
	file := flag.String("f", "stock.json", "stock file")
	id := flag.String("id", "cli", "order ID published with OrderPlaced")
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}
	if err := run(*file, *id, flag.Arg(0), flag.Args()[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "shop:", err)
		os.Exit(1)
	}
}

func run(file, id, cmd string, args []string) error {
	switch cmd {
	case "init":
		st := stock.NewStock()
		st.Add("apple", 100)
		st.Add("pear", 50)
		return save(file, st)

	case "price":
		if len(args) != 2 {
			return errors.New("usage: price UNIT QTY")
		}
		unit, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("bad unit price %q", args[0])
		}
		qty, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("bad quantity %q", args[1])
		}
		fmt.Println(pricing.CalcPrice(unit, qty))
		return nil

	case "count":
		if len(args) != 1 {
			return errors.New("usage: count NAME")
		}
		st, err := load(file)
		if err != nil {
			return err
		}
		fmt.Println(st.Count(args[0]))
		return nil

	case "order":
		items, err := parseItems(args)
		if err != nil {
			return err
		}
		st, err := load(file)
		if err != nil {
			return err
		}
		total, err := newService(st).Place(id, items)
		if err != nil {
			return err
		}
		if err := save(file, st); err != nil {
			return err
		}
		fmt.Println("total", total)
		return nil
	}
	return fmt.Errorf("unknown command %q", cmd)
}

func newService(st *stock.Stock) *orders.Service {
	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) {
		for name, qty := range e.Items {
			_ = st.Remove(name, qty)
		}
	})
	return orders.New(st, bus)
}

// parseItems reads order lines written as NAME=QTY@PRICE, such as apple=2@0.5.
func parseItems(args []string) ([]orders.Item, error) {
	if len(args) == 0 {
		return nil, errors.New("usage: order NAME=QTY@PRICE ...")
	}
	items := make([]orders.Item, 0, len(args))
	for _, arg := range args {
		name, rest, ok := strings.Cut(arg, "=")
		qtyText, priceText, ok2 := strings.Cut(rest, "@")
		if !ok || !ok2 || name == "" {
			return nil, fmt.Errorf("bad order line %q, want NAME=QTY@PRICE", arg)
		}
		qty, err := strconv.Atoi(qtyText)
		if err != nil {
			return nil, fmt.Errorf("bad quantity in %q", arg)
		}
		price, err := strconv.ParseFloat(priceText, 64)
		if err != nil {
			return nil, fmt.Errorf("bad price in %q", arg)
		}
		items = append(items, orders.Item{Name: name, UnitPrice: price, Qty: qty})
	}
	return items, nil
}
