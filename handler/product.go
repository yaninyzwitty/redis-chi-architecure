package handler

import (
	"net/http"

	"github.com/yaninyzwitty/merch-crud-microservice-go/model"
)

type Order struct {
	Repo *model.Product
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Create an order"))

}
func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("List orders"))

}
func (o *Order) GetById(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("et by Id"))

}
func (o *Order) UpdateById(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Update by id"))

}
