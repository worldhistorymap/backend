package middleware

import "net/http"

type Middleware func(h http.HandlerFunc) http.HandlerFunc

type MiddlewareIn func(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc

type MiddlewareOut func(h http.HandlerFunc, middlewares ...Middleware)


var InChain MiddlewareIn = func(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for i := len(middlewares) - 1; i >=0; i-- {
		h = middlewares[i](h)
	}
	return h
}

var OutChain MiddlewareOut = func(h http.HandlerFunc, middlewares ...Middleware) {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
}

//Middleware Chains


