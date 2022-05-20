package ekko

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
)

var (
	routerHandler = "handler"
	routerPath    = "path"
)

type handler func(w http.ResponseWriter, r *http.Request, c RouterParams)

type RouterParams map[string]string

func (r RouterParams) GetParams(s string) string {
	return r[s]
}

type trieNode struct {
	isNext      map[string]*trieNode
	isEnd       bool
	nodeContext context.Context
}

type routerCtx struct {
	Handler handler
	params  RouterParams
}

type Router struct {
	routesGet  *trieNode
	routesPost *trieNode
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method
	// todo fmt.Println("[method]: ", method)
	// todo fmt.Println("[path]: ", path)

	switch method {
	case http.MethodGet:
		// todo log.Println("match GET")
		routerCtx := router.match("GET", path)
		// todo fmt.Printf("params: %v\n", routerCtx.params.GetParams("id")) // BUG
		if routerCtx.Handler != nil {
			routerCtx.Handler(w, r, routerCtx.params)
		} else {
			log.Println("handler not found")
		}
	case http.MethodPost:
		// todo log.Println("match POST")
		routerCtx := router.match("POST", path)
		fmt.Printf("params: %v\n", routerCtx.params.GetParams("id")) // BUG
		if routerCtx.Handler != nil {
			routerCtx.Handler(w, r, routerCtx.params)
		} else {
			log.Println("handler not found")
		}
	}
}

// match 路由匹配
// return handler and router params.
func (router *Router) match(method string, requestPath string) routerCtx {
	var trie *trieNode
	switch method {
	case http.MethodGet:
		trie = router.routesGet
	case http.MethodPost:
		trie = router.routesPost
	}
	pathSlice := strings.Split(requestPath, "/")
	cur := trie
	for i, nodeName := range pathSlice { // if path = "/dave/home" match => "/:id/home"
		// 优先匹配静态路径
		if cur.isNext[nodeName] == nil {
			// 动态路径
			cur = cur.isNext["*"]
		} else {
			// 静态路径
			cur = cur.isNext[nodeName]
		}

		if cur == nil {
			log.Println("not find path, " + requestPath)
			return routerCtx{} // BUG: there have a error, maybe.
		} else if i == len(pathSlice)-1 && cur.isEnd {
			handler, OK := cur.nodeContext.Value(routerHandler).(handler)
			if OK && handler != nil {
				// todo log.Println("match path, ", requestPath)
				return routerCtx{
					Handler: handler,
					params:  router.pathParse(cur.nodeContext.Value(routerPath).(string), requestPath),
				}
			} else {
				return routerCtx{}
			}
		}
	}
	return routerCtx{}
}

// pathParse parse dynamic router path's params
// such as this path "/dave/home" match this router "/:id/home"
// that the variable `id` = "dave"
func (router *Router) pathParse(routerPath string, path string) RouterParams {
	routerPathSlice := strings.Split(routerPath, "/")
	pathSlice := strings.Split(path, "/")
	if len(routerPathSlice) != len(pathSlice) {
		panic("parse failed; the router path len not equal the request path len")
	}
	params := RouterParams{}
	for i := 0; i < len(pathSlice); i++ {
		if routerPathSlice[i] == pathSlice[i] {
			continue
		}
		params[strings.Replace(routerPathSlice[i], ":", "", 1)] = pathSlice[i]
	}
	return params
}

func (router *Router) insert(method string, path string, handler handler) {
	var trie *trieNode
	switch method {
	case http.MethodGet:
		trie = router.routesGet
	case http.MethodPost:
		trie = router.routesPost
	}
	pathSlice := strings.Split(path, "/")
	// todo log.Printf("%v", pathSlice)
	if trie == nil {
		panic("insert router error, trie is nil")
	}
	cur := trie
	for i, nodeName := range pathSlice {
		// 替换为通配符
		// BUG: it's not judge that it's legal the subRouterPath or not, such as ":id:", there is two ":" symbol.
		if strings.HasPrefix(nodeName, ":") {
			nodeName = "*"
		}
		if cur.isNext[nodeName] == nil {
			newNode := trieNode{
				isNext: make(map[string]*trieNode),
			}
			if i == len(pathSlice)-1 {
				ctx := context.WithValue(context.Background(), routerHandler, handler)
				ctx = context.WithValue(ctx, routerPath, path)
				newNode.nodeContext = ctx
				newNode.isEnd = true
			}
			cur.isNext[nodeName] = &newNode
		}
		cur = cur.isNext[nodeName]
	}
}

func (router *Router) Get(path string, handler handler) {
	// 先判断path是否合法: "", "home/dave" 等等这些都是不合法的, 必须以"/"开头
	if len(path) == 0 || !strings.HasPrefix(path, "/") {
		panic("This router path not allowed.")
	}

	router.insert("GET", path, handler)
}

func (router *Router) Post(path string, handler handler) {
	// 先判断path是否合法: "", "home/dave" 等等这些都是不合法的, 必须以"/"开头
	if len(path) == 0 || !strings.HasPrefix(path, "/") {
		panic("This router path not allowed.")
	}
	router.insert("POST", path, handler)
}

func NewRouter() *Router {
	return &Router{
		routesGet: &trieNode{
			isNext: map[string]*trieNode{},
		},
		routesPost: &trieNode{
			isNext: map[string]*trieNode{},
		},
	}
}
