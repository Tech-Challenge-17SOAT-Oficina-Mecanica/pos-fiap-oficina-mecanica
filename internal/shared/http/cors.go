package http

import stdhttp "net/http"

func CORS(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(writer stdhttp.ResponseWriter, request *stdhttp.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, If-Match")
		if request.Method == stdhttp.MethodOptions {
			writer.WriteHeader(stdhttp.StatusNoContent)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
