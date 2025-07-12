package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "available",
		"env":     app.config.env,
		"version": version,
	}

	err := app.writeJson(w, http.StatusOK, data, nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}

	// // use json.NewEncoder; 但是不方便在Encoder出错时做对应处理，可以借助bytes.Buffer临时存储好方便判断，但是又麻烦了，还是Marshal方便
	// err = json.NewEncoder(w).Encode(data)
	// if err != nil {
	// 	app.logger.Println(err)
	// 	http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	// }
}
