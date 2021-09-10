package main

import (
	"fmt"
	"os"
	"io"
	"net/http"
	"log"
	"github.com/gorilla/mux"
	"encoding/json"
	"github.com/google/uuid"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
)

type apiEndpoints struct{
	Domain string
	Endpoint string
}

func tokenHandler(s *server.Server) http.HandlerFunc{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      s.HandleTokenRequest(w,r)
   })
}

func credentialsHandler(c *store.ClientStore) http.HandlerFunc{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      clientId := uuid.New().String()[:8]
      clientSecret := uuid.New().String()[:8]
      err := c.Set(clientId, &models.Client{
         ID:     clientId,
         Secret: clientSecret,
         Domain: "http://localhost:8081",
      })
      if err != nil {
         fmt.Println(err.Error())
      }

      w.Header().Set("Content-Type", "application/json")
      json.NewEncoder(w).Encode(map[string]string{"CLIENT_ID": clientId, "CLIENT_SECRET": clientSecret})
   })
}

func validateToken(f http.HandlerFunc, srv *server.Server) http.HandlerFunc {
   return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      _, err := srv.ValidationBearerToken(r)
      if err != nil {
         http.Error(w, err.Error(), http.StatusBadRequest)
         return
      }

      f.ServeHTTP(w, r)
   })
}

func getoAuthConfig() (*store.ClientStore,*server.Server) {
	manager:=manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	clientStore:=store.NewClientStore()

	manager.MapClientStorage(clientStore)
	manager.SetRefreshTokenCfg(manage.DefaultRefreshTokenCfg)

	srv:=server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
    srv.SetClientInfoHandler(server.ClientFormHandler)

    srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
      log.Println("Internal Error:", err.Error())
      return
    })

    srv.SetResponseErrorHandler(func(re *errors.Response) {
      log.Println("Response Error:", re.Error.Error())
    })

	return clientStore,srv
}

func redirectHandler(w http.ResponseWriter, r *http.Request){

	endpoints:=make(map[string]apiEndpoints)
	endpoints["test"]=apiEndpoints{"http://localhost:8080/","test"}
	endpoints["protected"]=apiEndpoints{"http://localhost:8080/","protected"}

	fmt.Println("Redirect url called")
	fmt.Printf("\nRequest - %+v\n", r)
	params:=mux.Vars(r)
	fmt.Printf("\nParams - %+v\n", params)

	api:=endpoints[params["path"]]

	httpClient := http.Client{}
	reqUrl:=fmt.Sprintf("%s%s?access_token=%s",api.Domain,api.Endpoint,"abcd")
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)

	if err != nil {
		fmt.Fprintf(os.Stdout, "could not create HTTP request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}
	req.Header.Set("accept", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	fmt.Printf("\nResponse - %+v\n", res)
	defer res.Body.Close()

	b,err:=io.ReadAll(res.Body)
	if err!=nil{
		fmt.Println("error - %v", err)
	}
	w.Write(b)
}

func main(){
	fmt.Println("Hello, Token Server Started")

	clientStore,srv:=getoAuthConfig()
    
	router:=mux.NewRouter()
   router.HandleFunc("/token", tokenHandler(srv)).Methods("GET")
   router.HandleFunc("/credentials", credentialsHandler(clientStore)).Methods("GET")
   router.HandleFunc("/redirect/{path}", validateToken(redirectHandler,srv)).Methods("GET")
	log.Fatal(http.ListenAndServe(":8084",router))
}
