package main

import "net/http"

func (app *application) enableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// NOTES: Here is how you allow credentials and CORS on your backend server.
		//	remember that your have to set the proxy on your frontend to fool the browser that
		//	its making a request on the same domain & port number

		// To allow any http request (http://*), do this:
		// 		w.Header().Set("Access-Control-Allow-Origin", "http://*")

		// To specifically allow only requests from one domain and port eg localhost & port 3000,
		//		do this: w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")

		// We will make it secure (with SSL) when we get to production
		// Set CORS headers for all requests
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-CSRF-Token, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}

		// For non-OPTIONS requests, continue processing the request
		h.ServeHTTP(w, r)
	})
}

func (app *application) authRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _, err := app.auth.GetTokenFromHeaderAndVerify(w, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
