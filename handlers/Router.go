package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/JojiiOfficial/DataManagerServer/models"
	"github.com/jinzhu/gorm"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type handlerData struct {
	config *models.Config
	db     *gorm.DB
}

//Route for REST
type Route struct {
	Name        string
	Method      HTTPMethod
	Pattern     string
	HandlerFunc RouteFunction
	HandlerType requestType
}

//HTTPMethod http method. GET, POST, DELETE, HEADER, etc...
type HTTPMethod string

//HTTP methods
const (
	GetMethod    HTTPMethod = "GET"
	POSTMethod   HTTPMethod = "POST"
	DeleteMethod HTTPMethod = "DELETE"
)

type requestType uint8

const (
	defaultRequest requestType = iota
	sessionRequest
	optionalTokenRequest
)

//Routes all REST routes
type Routes []Route

//RouteFunction function for handling a route
type RouteFunction func(handlerData, http.ResponseWriter, *http.Request)

//Routes
var (
	routes = Routes{
		//Ping
		Route{
			Name:        "ping",
			Pattern:     "/ping",
			Method:      POSTMethod,
			HandlerFunc: Ping,
			HandlerType: defaultRequest,
		},
		//User
		Route{
			Name:    "login",
			Pattern: "/user/login",
			Method:  POSTMethod,
			//HandlerFunc: Login,
			HandlerType: defaultRequest,
		},
		Route{
			Name:    "register",
			Pattern: "/user/create",
			Method:  POSTMethod,
			//HandlerFunc: Register,
			HandlerType: defaultRequest,
		},

		//Files
		Route{
			Name:        "upload",
			Pattern:     "/file/upload",
			Method:      POSTMethod,
			HandlerFunc: UploadfileHandler,
			HandlerType: defaultRequest,
		},
		Route{
			Name:        "list files",
			Pattern:     "/file/list",
			Method:      POSTMethod,
			HandlerFunc: ListFilesHandler,
			HandlerType: defaultRequest,
		},
	}
)

//NewRouter create new router
func NewRouter(config *models.Config, db *gorm.DB) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(string(route.Method)).
			Path(route.Pattern).
			Name(route.Name).
			Handler(RouteHandler(route.HandlerType, &handlerData{
				config: config,
				db:     db,
			}, route.HandlerFunc, route.Name))
	}
	return router
}

//RouteHandler logs stuff
func RouteHandler(requestType requestType, handlerData *handlerData, inner RouteFunction, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("[%s] %s\n", r.Method, name)

		start := time.Now()

		if validateHeader(handlerData.config, w, r) {
			return
		}

		//Validate request by requestType
		//if !requestType.validate(db, handlerData, r, w) {
		//return
		//}

		//Process request
		inner(*handlerData, w, r)

		//Print duration of processing
		printProcessingDuration(start)
	})
}

//Return false on error
func (requestType requestType) validate(handlerData *handlerData, r *http.Request, w http.ResponseWriter) bool {
	switch requestType {
	case sessionRequest:
		{
			authHandler := NewAuthHandler(r)
			_ = authHandler
			//Do sth with the authhandler

		}
	}

	return true
}

//Prints the duration of handling the function
func printProcessingDuration(startTime time.Time) {
	dur := time.Since(startTime)

	if dur < 1500*time.Millisecond {
		log.Debugf("Duration: %s\n", dur.String())
	} else if dur > 1500*time.Millisecond {
		log.Warningf("Duration: %s\n", dur.String())
	}
}

//Return true on error
func validateHeader(config *models.Config, w http.ResponseWriter, r *http.Request) bool {
	headerSize := getHeaderSize(r.Header)

	//Send error if header are too big. MaxHeaderLength is stored in b
	if headerSize > uint32(config.Webserver.MaxHeaderLength) {
		//Send error response
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		fmt.Fprint(w, "413 request too large")

		log.Warnf("Got request with %db headers. Maximum allowed are %db\n", headerSize, config.Webserver.MaxHeaderLength)
		return true
	}

	return false
}
