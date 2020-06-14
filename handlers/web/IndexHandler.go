package web

import "net/http"

//FavIconHandler handle favicon
func FavIconHandler(handlerData HandlerData, w http.ResponseWriter, r *http.Request) error {
	/* #nosec */
	serveStaticFile(handlerData.Config, FavIconFile, w)

	return nil
}

//IndexPageHandler show index/main page
func IndexPageHandler(handlerData HandlerData, w http.ResponseWriter, r *http.Request) error {
	handleBrowserServeError(
		//Try to serve index file
		serveStaticFile(handlerData.Config, IndexFile, w),
		handlerData, w, r)

	return nil
}
