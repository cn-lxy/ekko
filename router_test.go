package ekko

import (
	"fmt"
	"net/http"
	"testing"
)

func TestRouter_Get(t *testing.T) {
	// next three test will fail.
	t.Run("testGet1", func(t *testing.T) {
		router := Router{}
		router.Get("home", func(w http.ResponseWriter, r *http.Request, p RouterParams) {
			_, _ = w.Write([]byte("Hello"))
		})
	})
	t.Run("testGet2", func(t *testing.T) {
		router := Router{}
		router.Get("", func(w http.ResponseWriter, r *http.Request, p RouterParams) {
			_, _ = w.Write([]byte("Hello"))
		})
	})
	t.Run("testGet3", func(t *testing.T) {
		router := Router{}
		router.Get("home/dave", func(w http.ResponseWriter, r *http.Request, p RouterParams) {
			_, _ = w.Write([]byte("Hello"))
		})
	})
	// this test will success.
	t.Run("testGet4", func(t *testing.T) {
		router := Router{}
		router.Get("/home", func(w http.ResponseWriter, r *http.Request, p RouterParams) {
			_, _ = w.Write([]byte("Hello"))
		})
		fmt.Println("root node child len: ", len(router.routesGet.isNext))
	})
}

func TestRunning(t *testing.T) {
	router := NewRouter()
	router.Get("/hello/:name", func(w http.ResponseWriter, r *http.Request, c RouterParams) {
		name := c.GetParams("name")
		_, _ = w.Write([]byte("hello, " + name))
	})

	err := http.ListenAndServe("localhost:8080", router)
	if err != nil {
		return
	}
}
