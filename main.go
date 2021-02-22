package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"github.com/go-chi/cors"
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

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins:   []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Access-Control-Allow-Origin", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	  }))

	r.Mount("/print", printerResource{p: ps}.Routes())

	http.ListenAndServe(":3000", r)
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
	OrderInformation []OrderInformation `json:"orderInformation"`
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
	AnswerNumber int    `json:"answerNumber"`
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
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// o := Order{
		// 	ID: "testorderID",
		// 	OrderInformation: []OrderInformation{
		// 		{Question: "this is a question?", AnswerString: "test"},
		// 		{Question: "this is a question as well?", AnswerNumber: 2},
		// 	},
		// }
		pr.p.AddToPrintQueue(o)
		w.WriteHeader(http.StatusOK)


		return
	})
	return r
}
