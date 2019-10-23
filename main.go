package main

import "net/http"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})
	println("Start server at port 8080...")
	http.ListenAndServe(":8080", nil)
}
