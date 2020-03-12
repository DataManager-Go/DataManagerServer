package web

import "net/http"

//FavIconHandler handle favicon
func FavIconHandler(handlerData HandlerData, w http.ResponseWriter, r *http.Request) {
	serveStaticFile(handlerData.Config, FavIconFile, w)
}

//IndexPageHandler show index/main page
func IndexPageHandler(handlerData HandlerData, w http.ResponseWriter, r *http.Request) {
	handleBrowserServeError(
		//Try to serve index file
		serveStaticFile(handlerData.Config, IndexFile, w),
		handlerData, w, r)
}
