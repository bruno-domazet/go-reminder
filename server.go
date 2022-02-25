package remme

import (
	"http"
	"net/http"
)

func server() {
	http.HandleFunc("/", func(res http.ResponseWriter, req http.Request) {

	})

}
