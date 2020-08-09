package main

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type payee struct {
	bsb, acc string
}

var payees = map[string]payee{
	"Sample Foods":                 {"000-000", "000000000"},
}

func main() {
	f, err := os.Open("bills.htm")
	if err != nil {
		log.Fatal(err)
	}
	doc, err := html.Parse(f)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}
	t := find(doc, "tbody")
	if t == nil {
		log.Fatal("not found")
	}
	csvFilename := "out.csv"
	f, err = os.Create(csvFilename)
	if err != nil {
		log.Fatal(err)
	}
	extract(t, csv.NewWriter(f))
	f.Close()
	log.Printf("wrote to %s\n", csvFilename)
}

func extract(tb *html.Node, w *csv.Writer) {
	for r := tb.FirstChild; r != nil; r = r.NextSibling {
		var row []string
		for td := r.FirstChild; td != nil; td = td.NextSibling {
			if td.Data != "td" {
				continue
			}
			row = append(row, contentOf(td.FirstChild))
		}
		bill, inv, name, amt := row[0], row[1], row[2], row[5]
		f, err := strconv.ParseFloat(amt, 64)
		if err != nil || f <= 0 || f > 1000 {
			log.Printf("weird amount %s in %s / %s", amt, bill, inv)
			continue
		}
		payee, ok := payees[name]
		if !ok {
			log.Printf("unknown payee %q", name)
			continue
		}
		desc := inv
		if desc == "" {
			desc = bill
		}
		bsb := strings.Replace(payee.bsb, "-", "", 1)
		record := []string{bsb, payee.acc, name, desc, amt}
		log.Println(record)
		w.Write(record)
	}
	w.Flush()
}

func contentOf(n *html.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == html.TextNode {
		return n.Data
	}
	if s := contentOf(n.FirstChild); s != "" {
		return s
	}
	if s := contentOf(n.NextSibling); s != "" {
		return s
	}
	return "stuff"
}

func find(n *html.Node, tag string) *html.Node {
	if n == nil {
		return nil
	}
	if n.Type != html.TextNode {
		if n.Data == tag {
			return n
		}
	}
	if f := find(n.FirstChild, tag); f != nil {
		return f
	}
	if f := find(n.NextSibling, tag); f != nil {
		return f
	}
	return nil
}
