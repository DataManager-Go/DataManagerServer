package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/DataManager-Go/DataManagerServer/handlers/web"
	"github.com/DataManager-Go/DataManagerServer/models"

	"github.com/JojiiOfficial/gaw"
	"gorm.io/gorm"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Route for REST
type Route struct {
	Name        string
	Method      HTTPMethod
	Pattern     string
	HandlerFunc RouteFunction
	HandlerType requestType
}

// HTTPMethod http method. GET, POST, DELETE, HEADER, etc...
type HTTPMethod string

// HTTP methods
const (
	GetMethod    HTTPMethod = "GET"
	POSTMethod   HTTPMethod = "POST"
	PUTMethod    HTTPMethod = "PUT"
	DeleteMethod HTTPMethod = "DELETE"
)

type requestType uint8

const (
	defaultRequest requestType = iota
	sessionRequest
	optionalTokenRequest
)

// Routes all REST routes
type Routes []Route

// RouteFunction function for handling a route
type RouteFunction func(web.HandlerData, http.ResponseWriter, *http.Request) error

// Routes
var (
	routes = Routes{
		// Ping
		Route{
			Name:        "ping",
			Pattern:     "/ping",
			Method:      POSTMethod,
			HandlerFunc: Ping,
			HandlerType: defaultRequest,
		},
		// User
		Route{
			Name:        "login",
			Pattern:     "/user/login",
			Method:      POSTMethod,
			HandlerFunc: Login,
			HandlerType: defaultRequest,
		},
		Route{
			Name:        "register",
			Pattern:     "/user/register",
			Method:      POSTMethod,
			HandlerFunc: Register,
			HandlerType: defaultRequest,
		},

		// Files
		Route{
			Name:        "upload",
			Pattern:     "/upload/file",
			Method:      PUTMethod,
			HandlerFunc: UploadfileHandler,
			HandlerType: sessionRequest,
		},
		Route{
			Name:        "list files",
			Pattern:     "/files",
			Method:      POSTMethod,
			HandlerFunc: ListFilesHandler,
			HandlerType: sessionRequest,
		},
		Route{
			Name:        "fileaction",
			Pattern:     "/file/{action}",
			Method:      POSTMethod,
			HandlerFunc: FileHandler,
			HandlerType: sessionRequest,
		},

		// Preview
		Route{
			Name:        "preview",
			Pattern:     "/preview/{fileID}",
			HandlerFunc: web.PrevievFileHandler,
			HandlerType: defaultRequest,
			Method:      GetMethod,
		},
		Route{
			Name:        "raw file",
			Pattern:     "/preview/raw/{fileID}",
			HandlerFunc: web.RawFileHandler,
			HandlerType: defaultRequest,
			Method:      GetMethod,
		},

		// Attribute
		Route{
			Name:        "Attribute",
			Pattern:     "/attribute/{attribute}/{action}",
			Method:      POSTMethod,
			HandlerFunc: AttributeHandler,
			HandlerType: sessionRequest,
		},
		Route{
			Name:        "User attribute data",
			Pattern:     "/attributes",
			Method:      POSTMethod,
			HandlerFunc: UserAttributeHandler,
			HandlerType: sessionRequest,
		},

		// Namespace
		Route{
			Name:        "Namespace",
			Pattern:     "/namespace/{action}",
			Method:      POSTMethod,
			HandlerFunc: NamespaceActionHandler,
			HandlerType: sessionRequest,
		},
		Route{
			Name:        "Namespace list",
			Pattern:     "/namespaces",
			Method:      POSTMethod,
			HandlerFunc: NamespaceListHandler,
			HandlerType: sessionRequest,
		},
		// TODO add stats
	}
)

// NewRouter create new router and its required components
func NewRouter(config *models.Config, db *gorm.DB) *mux.Router {
	handlerData := web.HandlerData{
		Config: config,
		Db:     db,
	}

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(string(route.Method)).
			Path(route.Pattern).
			Name(route.Name).
			Handler(RouteHandler(route.HandlerType, &handlerData, route.HandlerFunc, route.Name))
	}

	// Adding custom routes
	addCustomRoutes(router, &handlerData)

	return router
}

// Add custom web-routes
func addCustomRoutes(router *mux.Router, handlerData *web.HandlerData) {
	// 404 Handler
	router.NotFoundHandler = RouteHandler(defaultRequest, handlerData, web.NotFoundHandler, "not found")

	// Index routes
	router.Handle("/", RouteHandler(defaultRequest, handlerData, web.IndexPageHandler, "index"))

	//Favicon
	router.Handle("/favicon.ico", RouteHandler(defaultRequest, handlerData, web.FavIconHandler, ""))

	// Serve static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./html/static"))))
}

// RouteHandler logs stuff
func RouteHandler(requestType requestType, handlerData *web.HandlerData, inner RouteFunction, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := r.Body.Close(); err != nil {
				log.Info(err)
			}
		}()

		// Only debug routes which have a names
		needDebug := len(name) > 0
		if needDebug {
			log.Infof("[%s] %s\n", r.Method, name)
		}

		start := time.Now()

		if validateHeader(handlerData.Config, w, r) {
			return
		}

		// Validate request by requestType
		if !requestType.validate(handlerData, r, w) {
			return
		}

		// Process request and handle its error
		if err := inner(*handlerData, w, r); err != nil {
			if e, ok := err.(*RequestError); ok {
				// Send error response to user
				sendResponse(w, models.ResponseError, e.String(), "", e.ResponseCode)

				// Log user errors on debugmode
				if log.GetLevel() == log.DebugLevel {
					LogError(err)
				}
			} else {
				sendServerError(w)
				log.Error(err)
			}
		}

		// Print duration of processing
		if needDebug {
			printProcessingDuration(start)
		}
	})
}

// Validate the given request based on its requestType
// if validate returns false, the request has to abort
func (requestType requestType) validate(handlerData *web.HandlerData, r *http.Request, w http.ResponseWriter) bool {
	switch requestType {
	case sessionRequest:
		{
			// SessionRequest requires a valid session token
			// which identifies a specific user

			// Create an AuthHandler using r
			authHandler := NewAuthHandler(r)

			// Check Token validity by its length
			if len(authHandler.GetBearer()) != 64 {
				log.Error("Invalid token len %d", len(authHandler.GetBearer()))
				sendResponse(w, models.ResponseError, "Invalid token", http.StatusUnauthorized)
				return false
			}

			// Try to retrieve the user by the requestToken
			user, err := models.GetUserFromSession(handlerData.Db, authHandler.GetBearer())
			if LogError(err) || user == nil {
				if user == nil && err == nil {
					log.Error("Can't get user")
				}

				sendResponse(w, models.ResponseError, "Invalid token", http.StatusUnauthorized)
				return false
			}

			// Set the handlerData.User which is
			// required by all endpoint functions
			handlerData.User = user
		}
	}

	return true
}

// Prints the duration of handling the function
func printProcessingDuration(startTime time.Time) {
	dur := time.Since(startTime)

	if dur < 1500*time.Millisecond {
		log.Debugf("Duration: %s\n", dur.String())
	} else if dur > 1500*time.Millisecond {
		log.Warningf("Duration: %s\n", dur.String())
	}
}

// Return true on error
func validateHeader(config *models.Config, w http.ResponseWriter, r *http.Request) bool {
	headerSize := gaw.GetHeaderSize(r.Header)

	// Send error if header are too big. MaxHeaderLength is stored in b
	if headerSize > uint32(config.Webserver.MaxHeaderLength) {
		// Send error response
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		fmt.Fprint(w, "413 request too large")

		log.Warnf("Got request with %db headers. Maximum allowed are %db\n", headerSize, config.Webserver.MaxHeaderLength)
		return true
	}

	return false
}
