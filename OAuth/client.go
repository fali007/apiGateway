package main

import (
	"fmt"
	"net/http"
	"log"
	"github.com/gorilla/mux"
)

type authBook struct{
	clientId string
	secretKey string
	authUrl string
	redirectUrl string
}

func testhandler(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("Hello, This is not secured"))
}

func protectedhandler(w http.ResponseWriter, r *http.Request){
	fmt.Println("protectedhandler is called")
	w.Write([]byte("Hello, this is secured. Please dont try to break in.. I would have to kill you.."))
}

func main(){
	fmt.Println("Hello, Server Started")
    
	router:=mux.NewRouter()
	router.HandleFunc("/test",testhandler).Methods("GET")
	router.HandleFunc("/protected",protectedhandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080",router))
}
