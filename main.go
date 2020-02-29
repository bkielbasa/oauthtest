package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

func main() {
	conf := &oauth2.Config{
		ClientID:     envOrPanic("CLIENT_ID"),
		ClientSecret: envOrPanic("CLIENT_SECRET"),
		RedirectURL:  "http://localhost/oauth",
		Scopes:       []string{"email"},
		Endpoint:     endpoints.Facebook,
	}

	port := 80
	handler := oauthHandler{conf}

	http.HandleFunc("/", handler.auth)
	http.HandleFunc("/oauth", handler.handleCode)
	log.Printf("application started on port %d", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

type oauthHandler struct {
	conf *oauth2.Config
}

func (h oauthHandler) auth(w http.ResponseWriter, r *http.Request) {
	url := h.conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	body := fmt.Sprintf(`<p>To authenticate please <a href="%v">click here</a></p>`, url)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, body)
}

func (h oauthHandler) handleCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := r.URL.Query()["code"][0]
	tok, err := h.conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	client := h.conf.Client(ctx, tok)
	resp, err := client.Get("https://graph.facebook.com/me?access_token=" + url.QueryEscape(tok.AccessToken))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w, string(body))
}

func envOrPanic(name string) string {
	v := os.Getenv(name)
	if v == "" {
		panic(fmt.Sprintf("the env variable `%s` does not exist", name))
	}

	return v
}
