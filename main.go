package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/spf13/viper"
)

func main() {

	configSetup()

	pa := viper.GetString("server.printer_address")
	pp := viper.GetInt("server.printer_port")

	ps, err := NewPrinterService(pa, pp)
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Access-Control-Allow-Origin", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Mount("/print", printerResource{p: ps}.Routes())

	sp := viper.GetInt("server.port")
	sps := ":" + strconv.Itoa(sp)

	fmt.Printf("Server listening on port: %d\n", sp)

	n := time.Now()
	warmup := Order{
		ID: "123456",
		OrderedItems: []MenuItem{
			{Name: "Warmup1", Price: 100},
			{Name: "Warmup2", Price: 200},
		},
		SubmittedTime: &n,
	}

	err = ps.AddToPrintQueue(warmup)
	if err != nil {
		fmt.Printf("Printer warmup print failed")
		log.Fatal(err)
	}
	http.ListenAndServe(sps, r)
}

func printHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		return
	}
}

type Order struct {
	ID               string             `json:"id"`
	ActivityID       string             `json:"activityID"`
	VisitorID        string             `json:"visitorID"`
	OrganisationID   string             `json:"organizationID"`
	Type             string             `json:"orderType"`
	OrderInformation []OrderInformation `json:"deliveryInformation"`
	OrderedItems     []MenuItem         `json:"items"`
	SubmittedTime    *time.Time         `json:"orderSubmittedTime"`
}

type MenuItem struct {
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type OrderInformation struct {
	Question     string `json:"question"`
	AnswerString string `json:"answerString"`
	AnswerNumber int    `json:"answerInt"`
}

type IPrinter interface {
	AddToPrintQueue(Order) error
}

type printerResource struct {
	p IPrinter
}

func (pr printerResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {

		o := Order{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&o)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(o.ID) > 4 {
			fmt.Printf("New Order: %s \n", o.ID[:4])
		} else {
			fmt.Printf("New Order: %s \n", o.ID)
		}

		pr.p.AddToPrintQueue(o)
		w.WriteHeader(http.StatusOK)

		return
	})
	return r
}
