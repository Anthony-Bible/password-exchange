// forms.go
package main

import (
    "html/template"
    "net/http"
    "log"
    "github.com/bmizerany/pat"
    "fmt"
)

// type Message struct {
//     Email   string
//     Subject string
//     Message string
// }

func main() {
  mux := pat.New()
  mux.Get("/", http.HandlerFunc(home))
  mux.Post("/", http.HandlerFunc(send))
  mux.Get("/confirmation", http.HandlerFunc(confirmation))

  log.Println("Listening...")
  err := http.ListenAndServe(":3000", mux)
  if err != nil {
    log.Fatal(err)
  }
}

func home(w http.ResponseWriter, r *http.Request) {
  render(w, "templates/home.html", nil)
}
func send(w http.ResponseWriter, r *http.Request) {
  // Step 1: Validate form
  // Step 2: Send message in an email
  // Step 3: Redirect to confirmation page
  msg := &Message{
		Email:   r.PostFormValue("email"),
		Content: r.PostFormValue("content"),
	}

	if msg.Validate() == false {
		render(w, "templates/home.html", msg)
		return
	}
  string, randmError := GenerateRandomString(32)
    if randmError != nil {
      log.Fatal(randmError)
    }
  siteHost := GetViperVariable("host")
  fmt.Println(siteHost)
  msg.Content = msg.Content +  "\n" siteHost + "encrypt/" + string
	if err := msg.Deliver(); err != nil {
		log.Println(err)
		http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/confirmation", http.StatusSeeOther)

}
func confirmation(w http.ResponseWriter, r *http.Request) {
  render(w, "templates/confirmation.html", nil)
}
func render(w http.ResponseWriter, filename string, data interface{}) {
  tmpl, err := template.ParseFiles(filename)
  if err != nil {
    log.Println(err)
    http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
  }

  if err := tmpl.Execute(w, data); err != nil {
    log.Println(err)
    http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
  }
}

