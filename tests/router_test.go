package tests

import (
	"testing"

	"github.com/Ruannilton/go-rate-limiter/internal"
)

func assertRouteMatch(t *testing.T, router *internal.Router, path string, expectedHandler internal.AlgorithmHandler) {
	t.Helper()
	handler, found := router.GetRoute(path)
	if !found {
		t.Errorf("expected to find route for path %q, but did not", path)
		return
	}
	if handler != expectedHandler {
		t.Errorf("path %q matched wrong handler", path)
	}
}

func assertRouteNotFound(t *testing.T, router *internal.Router, path string) {
	t.Helper()
	_, found := router.GetRoute(path)
	if found {
		t.Errorf("expected not to find route for path %q, but did", path)
	}
}

func TestRouter_StaticRoutes(t *testing.T) {
	router := internal.NewRouter()

	handler1 := internal.NewHandler(internal.NewFixedWindowMemoryStorage())
	handler2 := internal.NewHandler(internal.NewFixedWindowMemoryStorage())
	handler3 := internal.NewHandler(internal.NewFixedWindowMemoryStorage())

	router.AddRoute("/a/b/c", handler1)
	router.AddRoute("/a/b/d", handler2)
	router.AddRoute("/", handler3)

	assertRouteMatch(t, router, "/a/b/c", handler1)
	assertRouteMatch(t, router, "/a/b/d", handler2)
	assertRouteMatch(t, router, "/", handler3)
	assertRouteNotFound(t, router, "/a/b")
	assertRouteNotFound(t, router, "/a/b/c/d")
	assertRouteNotFound(t, router, "/x/y/z")
}

func TestRouter_VarRoutes(t *testing.T) {
	router := internal.NewRouter()

	handler1 := internal.NewHandler(internal.NewLeakyBucketMemoryStorage())
	handler2 := internal.NewHandler(internal.NewLeakyBucketMemoryStorage())

	router.AddRoute("/users/:id", handler1)
	router.AddRoute("/users/:id/profile", handler2)

	assertRouteMatch(t, router, "/users/123", handler1)
	assertRouteMatch(t, router, "/users/abc", handler1)
	assertRouteMatch(t, router, "/users/123/profile", handler2)
	assertRouteNotFound(t, router, "/users")
}

func TestRouter_WildcardRoutes(t *testing.T) {
	router := internal.NewRouter()

	handler1 := internal.NewHandler(internal.NewTokenBucketMemoryStorage())
	router.AddRoute("/files/*", handler1)

	assertRouteMatch(t, router, "/files/image.jpg", handler1)
	assertRouteMatch(t, router, "/files/document.pdf", handler1)
	assertRouteNotFound(t, router, "/files")
	assertRouteNotFound(t, router, "/files/images/jpeg")
}

func TestRouter_Precedence(t *testing.T) {
	router := internal.NewRouter()

	staticHandler := internal.NewHandler(internal.NewFixedWindowMemoryStorage())
	varHandler := internal.NewHandler(internal.NewFixedWindowMemoryStorage())
	wildcardHandler := internal.NewHandler(internal.NewFixedWindowMemoryStorage())

	router.AddRoute("/users/static", staticHandler)
	router.AddRoute("/users/:id", varHandler)
	router.AddRoute("/users/*", wildcardHandler)

	assertRouteMatch(t, router, "/users/static", staticHandler)
	assertRouteMatch(t, router, "/users/anything", varHandler)

	postVarHandler := internal.NewHandler(internal.NewLeakyBucketMemoryStorage())
	postWildcardHandler := internal.NewHandler(internal.NewLeakyBucketMemoryStorage())

	router.AddRoute("/posts/:id/edit", postVarHandler)
	router.AddRoute("/posts/*/edit", postWildcardHandler)
	assertRouteMatch(t, router, "/posts/123/edit", postVarHandler)
}

func TestRouter_Backtracking(t *testing.T) {
	router := internal.NewRouter()

	handler1 := internal.NewHandler(internal.NewFixedWindowMemoryStorage())
	handler2 := internal.NewHandler(internal.NewFixedWindowMemoryStorage())

	router.AddRoute("/a/b/c", handler1)
	router.AddRoute("/a/:id/d", handler2)

	assertRouteMatch(t, router, "/a/b/d", handler2)
	assertRouteNotFound(t, router, "/a/b/e")
}

func TestRouter_ComplexPrecedence(t *testing.T) {
	router := internal.NewRouter()

	staticHandler := internal.NewHandler(internal.NewFixedWindowMemoryStorage())
	varHandler := internal.NewHandler(internal.NewFixedWindowMemoryStorage())
	wildcardHandler := internal.NewHandler(internal.NewFixedWindowMemoryStorage())

	router.AddRoute("/api/v1/users/me", staticHandler)
	router.AddRoute("/api/v1/users/:id", varHandler)
	router.AddRoute("/api/v1/*", wildcardHandler)

	assertRouteMatch(t, router, "/api/v1/users/me", staticHandler)
	assertRouteMatch(t, router, "/api/v1/users/123", varHandler)
	assertRouteMatch(t, router, "/api/v1/posts", wildcardHandler)
	assertRouteNotFound(t, router, "/api/v2/users")
}
