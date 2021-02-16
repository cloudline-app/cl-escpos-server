package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {

	ps, err := NewPrinterService()
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Mount("/print", printerResource{p: ps}.Routes())

	http.ListenAndServe(":3000", r)
}

func printHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		return
	}
}

type Order struct {
	name string
}

type IPrinter interface {
	AddToPrintQueue(Order) error
}

type printerResource struct {
	p IPrinter
}

func (pr printerResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		pr.p.AddToPrintQueue(Order{name: "testing name"})
	})
	return r
}
