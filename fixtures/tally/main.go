// Command tally keeps item counts in a JSON stock file.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"example.com/tally/stock"
)

const usage = `usage: tally [-f file] <command>

commands:
  init               write a fresh stock file with 100 apples and 50 pears
  count NAME         print how many NAME are in stock
  add NAME QTY       put QTY units of NAME into stock
  order NAME QTY     sell QTY units of NAME
  discard NAME QTY   write off QTY spoiled units of NAME

flags:
`

func main() {
	file := flag.String("f", "stock.json", "stock file")
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}
	if err := run(*file, flag.Arg(0), flag.Args()[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "tally:", err)
		os.Exit(1)
	}
}

func run(file, cmd string, args []string) error {
	if cmd == "init" {
		s := stock.New()
		s.Add("apple", 100)
		s.Add("pear", 50)
		return s.Save(file)
	}

	s, err := stock.Load(file)
	if err != nil {
		return err
	}
	switch cmd {
	case "count":
		if len(args) != 1 {
			return errors.New("usage: count NAME")
		}
		fmt.Println(s.Count(args[0]))
		return nil

	case "add":
		name, qty, err := nameAndQty(cmd, args)
		if err != nil {
			return err
		}
		s.Add(name, qty)
		fmt.Printf("%s: %d in stock\n", name, s.Count(name))

	case "order":
		name, qty, err := nameAndQty(cmd, args)
		if err != nil {
			return err
		}
		if err := s.Remove(name, qty); err != nil {
			return fmt.Errorf("order %s: %w", name, err)
		}
		fmt.Printf("sold %d %s, %d left\n", qty, name, s.Count(name))

	case "discard":
		name, qty, err := nameAndQty(cmd, args)
		if err != nil {
			return err
		}
		if err := s.Remove(name, qty); err != nil {
			return fmt.Errorf("discard %s: %w", name, err)
		}
		fmt.Printf("discarded %d %s, %d left\n", qty, name, s.Count(name))

	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
	return s.Save(file)
}

func nameAndQty(cmd string, args []string) (string, int, error) {
	if len(args) != 2 {
		return "", 0, fmt.Errorf("usage: %s NAME QTY", cmd)
	}
	qty, err := strconv.Atoi(args[1])
	if err != nil || qty <= 0 {
		return "", 0, fmt.Errorf("quantity must be a positive integer, got %q", args[1])
	}
	return args[0], qty, nil
}
