package main

import (
	"fmt"
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

func testhandler(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("Hello, This is not secured"))
}

func protectedhandler(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("Hello, this is secured"))
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
         Domain: "http://localhost:8080",
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

func main(){
	fmt.Println("Hello, Server Started")

	clientStore,srv:=getoAuthConfig()
    
	router:=mux.NewRouter()
   router.HandleFunc("/token", tokenHandler(srv)).Methods("GET")
   router.HandleFunc("/credentials", credentialsHandler(clientStore)).Methods("GET")
	router.HandleFunc("/test",testhandler).Methods("GET")
	router.HandleFunc("/protected",validateToken(protectedhandler,srv)).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080",router))
}
