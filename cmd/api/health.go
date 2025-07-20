package main

import (
	"net/http"
	"time"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	time.Sleep(2 * time.Second)
	err := app.writeJson(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	// // use json.NewEncoder; 但是不方便在Encoder出错时做对应处理，可以借助bytes.Buffer临时存储好方便判断，但是又麻烦了，还是Marshal方便
	// err = json.NewEncoder(w).Encode(data)
	// if err != nil {
	// 	app.logger.Println(err)
	// 	http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	// }
}
