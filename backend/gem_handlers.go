package main

import (
	"net/http"
)

func ListGemsHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "List gems not implemented yet", http.StatusNotImplemented)
}

func GetGemHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Get gem not implemented yet", http.StatusNotImplemented)
}

func UpdateGemHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Update gem not implemented yet", http.StatusNotImplemented)
}

func DeleteGemHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Delete gem not implemented yet", http.StatusNotImplemented)
}

