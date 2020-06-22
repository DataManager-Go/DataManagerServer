package web

import (
	"net/http"
)

//NotFoundHandler 404 not found handler
func NotFoundHandler(handlerData HandlerData, w http.ResponseWriter, r *http.Request) error {
	/*
		log.Info("Not found: ", r.URL.Path)
		err := serveStaticFile(handlerData.Config, NotFoundFile, w)
		if err != nil {
			if os.IsNotExist(err) {
				log.Error("Can't find 404.html!")
				return nil
			}
		}
	*/

	w.Header().Add("Location", "/")
	w.WriteHeader(301)

	return nil
}
