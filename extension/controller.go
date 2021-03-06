package extension

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/qb0C80aE/clay/logging"
)

type regExpControllerPair struct {
	regExp     *regexp.Regexp
	controller Controller
}

var controllerList = []Controller{}
var pathControllerMap = map[string]*regExpControllerPair{}
var routerGroupMap = map[string]*gin.RouterGroup{}

// MethodSomething integer constants are used as the key value of HTTP methods in maps
const (
	MethodGet     = 1
	MethodPost    = 2
	MethodPut     = 3
	MethodDelete  = 4
	MethodPatch   = 5
	MethodOptions = 6
)

// URLSomething means URL type
const (
	URLSingle = 1
	URLMulti  = 2
)

// AcceptSomething contains the expression of string in Accept header
const (
	AcceptJSON     = "application/json"
	AcceptXYAML    = "application/x-yaml"
	AcceptTextYAML = "text/yaml"
	AcceptAll      = "*/*"
)

// AcceptCharsetSomething contains the expressin of string in Accept-Charset header
const (
	AcceptCharsetUTF8 = "utf-8"
)

// ContentTypeSomething contains the expression of string in Content-Type header
const (
	ContentTypeJSON              = "application/json"
	ContentTypeXYAML             = "application/x-yaml"
	ContentTypeTextYAML          = "text/yaml"
	ContentTypeMultipartFormData = "multipart/form-data"
	ContentTypeOctetStream       = "application/octet-stream"
)

var methodStringMap = map[int]string{
	MethodGet:     http.MethodGet,
	MethodPost:    http.MethodPost,
	MethodPut:     http.MethodPut,
	MethodDelete:  http.MethodDelete,
	MethodPatch:   http.MethodPatch,
	MethodOptions: http.MethodOptions,
}

var methodIntMap = map[string]int{
	http.MethodGet:     MethodGet,
	http.MethodPost:    MethodPost,
	http.MethodPut:     MethodPut,
	http.MethodDelete:  MethodDelete,
	http.MethodPatch:   MethodPatch,
	http.MethodOptions: MethodOptions,
}

// Controller is the interface what is mapped into specific urls and handles the requests from HTTP clients
// * GetModel returns its model
// * GetResourceName returns its resource/table name in REST/DB
// * GetResourceSingleURL builds a resource url what represents a single resource based on the argument
// * GetResourceMultiURL builds a resource url what represents multi resources based on the argument
// * GetRouteMap returns the map its key is resource url and its value is request handler of the controller
type Controller interface {
	GetModel() Model
	GetResourceName() (string, error)
	GetResourceSingleURL() (string, error)
	GetResourceMultiURL() (string, error)
	GetRouteMap() map[int]map[int]gin.HandlerFunc
}

// Binder is the interface what handles binding of input data
// * Bind binds input data to a container instance
type Binder interface {
	Bind(c *gin.Context, resourceName string) (interface{}, error)
}

// QueryCustomizer is the interface what handles query parameters used as Parameter struct in GetSingle and GetMulti
// * GetQueries returns query parameters
type QueryCustomizer interface {
	GetQueries(c *gin.Context) url.Values
}

// OutputHandler is the interface what handles outputs of the results from the logic to controllers
// * OutputError handles an error output
// * OutputGetSingle corresponds HTTP GET message and handles the output of a single result from logic classes
// * OutputGetMulti corresponds HTTP GET message and handles the output of multiple result from logic classes
// * OutputCreate corresponds HTTP POST message and handles the output of a single result from logic classes
// * OutputUpdate corresponds HTTP PUT message and handles the output of a single result from logic classes
// * OutputDelete corresponds HTTP DELETE message and handles the code result from logic classes
// * OutputPatch corresponds HTTP PATCH message and handles the output of a single result from logic classes
// * OutputGetOptions corresponds HTTP DELETE message and handles the code result from logic classes, as well as OutputDelete
type OutputHandler interface {
	OutputError(c *gin.Context, code int, err error)
	OutputGetSingle(c *gin.Context, code int, result interface{}, fields map[string]interface{})
	OutputGetMulti(c *gin.Context, code int, result interface{}, total int, countBeforePagination int, fields map[string]interface{})
	OutputCreate(c *gin.Context, code int, result interface{})
	OutputUpdate(c *gin.Context, code int, result interface{})
	OutputDelete(c *gin.Context, code int)
	OutputPatch(c *gin.Context, code int, result interface{})
	OutputGetOptions(c *gin.Context, code int)
}

// LookUpMethodName returns the HTTP method name string corresponded to the argument
func LookUpMethodName(method int) string {
	return methodStringMap[method]
}

// LookUpMethodInt returns the HTTP method int value corresponded to the argument
func LookUpMethodInt(method string) int {
	return methodIntMap[method]
}

// RegisterController registers a controller used in the router
func RegisterController(controller Controller) {
	controllerList = append(controllerList, controller)
}

// GetRegisteredControllerList returns the registered controllers
func GetRegisteredControllerList() []Controller {
	result := []Controller{}
	result = append(result, controllerList...)
	return result
}

// AssociateControllerWithPath associates a path and a controller
func AssociateControllerWithPath(path string, controller Controller) {
	pathElementList := strings.Split(strings.Trim(path, "/"), "/")

	newPathElementList := []string{}
	for _, pathElement := range pathElementList {
		if pathElement[:1] == ":" {
			newPathElementList = append(newPathElementList, "[^/]+")
		} else {
			newPathElementList = append(newPathElementList, pathElement)
		}
	}

	newPath := fmt.Sprintf("^%s$", strings.Join(newPathElementList, "/"))

	if oldPair, exists := pathControllerMap[newPath]; exists {
		if oldPair.controller != controller {
			logging.Logger().Criticalf("path %s is already mapped to another controller", newPath)
			os.Exit(1)
		}
	}

	newRegExp, err := regexp.Compile(newPath)
	if err != nil {
		logging.Logger().Critical(err.Error())
		os.Exit(1)
	}

	pathControllerMap[newPath] = &regExpControllerPair{
		regExp:     newRegExp,
		controller: controller,
	}
}

// GetAssociatedControllerWithPath returns the registered controller related to the given resource name
func GetAssociatedControllerWithPath(path string) (Controller, error) {
	path = strings.Trim(path, "/")
	for _, pair := range pathControllerMap {
		if pair.regExp.MatchString(path) {
			return pair.controller, nil
		}
	}
	logging.Logger().Critical("no controller is associated with %s", path)
	return nil, fmt.Errorf("no controller is associated with %s", path)
}

// RegisterRouterGroup registers a router group with a relative pth
func RegisterRouterGroup(relativePath string, routerGroup *gin.RouterGroup) error {
	if _, exists := routerGroupMap[relativePath]; exists {
		return fmt.Errorf("relative path %s is already registered", relativePath)
	}

	routerGroupMap[relativePath] = routerGroup

	return nil
}

// GetRegisteredRouterGroup returns a router group related to given path
func GetRegisteredRouterGroup(relativePath string) (*gin.RouterGroup, error) {
	routerGroup, exists := routerGroupMap[relativePath]
	if !exists {
		return nil, fmt.Errorf("relative path %s is not registered yet", relativePath)
	}

	return routerGroup, nil
}

// SetupController setups controllers
func SetupController(relativePath string, controllerList []Controller) error {
	routerGroup, err := GetRegisteredRouterGroup(relativePath)
	if err != nil {
		return err
	}

	methodFunctionMap := map[int]func(string, ...gin.HandlerFunc) gin.IRoutes{
		MethodGet:     routerGroup.GET,
		MethodPost:    routerGroup.POST,
		MethodPut:     routerGroup.PUT,
		MethodDelete:  routerGroup.DELETE,
		MethodPatch:   routerGroup.PATCH,
		MethodOptions: routerGroup.OPTIONS,
	}

	for _, controller := range controllerList {
		routeMap := controller.GetRouteMap()
		for method, routingFunction := range methodFunctionMap {
			routeList := routeMap[method]
			for pathType, handlerFunc := range routeList {
				switch pathType {
				case URLSingle:
					resourceSingleURL, err := controller.GetResourceSingleURL()
					if err != nil {
						logging.Logger().Critical(err.Error())
						return err
					}
					routingFunction(resourceSingleURL, handlerFunc)
					AssociateControllerWithPath(resourceSingleURL, controller)
				case URLMulti:
					resourceMultiURL, err := controller.GetResourceMultiURL()
					if err != nil {
						logging.Logger().Critical(err.Error())
						return err
					}
					routingFunction(resourceMultiURL, handlerFunc)
					AssociateControllerWithPath(resourceMultiURL, controller)
				default:
					logging.Logger().Criticalf("invalid url type: %d", pathType)
					return err
				}
			}
		}
	}

	return nil
}

// CreateParametersFromPathAndRoute creates parameters from actual access url path (like /test/1/generation) and routing url (like /test/:id/generation)
func CreateParametersFromPathAndRoute(path, route string) (gin.Params, error) {
	pathElements := strings.Split(strings.Trim(path, "/"), "/")
	routeElements := strings.Split(strings.Trim(route, "/"), "/")

	if len(pathElements) != len(routeElements) {
		logging.Logger().Debugf("the number of elements in path(%v) and route(%v) does not match", pathElements, routeElements)
		return gin.Params{}, fmt.Errorf("the number of elements in path(%v) and route(%v) does not match", pathElements, routeElements)
	}

	if len(pathElements) == 0 {
		logging.Logger().Debugf("the number of elements in path(%v) and route(%v) is zero", pathElements, routeElements)
		return gin.Params{}, fmt.Errorf("the number of elements in path(%v) and route(%v) is zero", pathElements, routeElements)
	}

	parameters := gin.Params{}
	for index, routeElement := range routeElements {
		if routeElement[:1] == ":" {
			parameter := gin.Param{
				Key:   routeElement[1:],
				Value: pathElements[index],
			}
			parameters = append(parameters, parameter)
		}
	}

	return parameters, nil
}

// DetermineResponseContentTypeFromAccept determines ContentType in the response header based on Accept header in the request
func DetermineResponseContentTypeFromAccept(acceptList []string) string {
	if len(acceptList) == 0 {
		return AcceptJSON
	}

	result := strings.Trim(acceptList[0], " ")
	switch result {
	case AcceptAll:
		return AcceptJSON
	default:
		return result
	}
}

// DetermineResponseMappingTagFromAccept determines the tag name based on Accept header in the request
func DetermineResponseMappingTagFromAccept(acceptList []string) string {
	switch DetermineResponseContentTypeFromAccept(acceptList) {
	case AcceptXYAML, AcceptTextYAML:
		return TagYAML
	default:
		return TagJSON
	}
}
