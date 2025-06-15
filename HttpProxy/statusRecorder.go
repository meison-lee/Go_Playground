package main

import "net/http"

type StatusRecorder struct {
	ResponseWriter http.ResponseWriter
	StatusCode     int
}

func (s *StatusRecorder) Header() http.Header {
	return s.ResponseWriter.Header()
}

func (s *StatusRecorder) Write(input []byte) (int, error) {
	return s.ResponseWriter.Write(input)
}

func (s *StatusRecorder) WriteHeader(statusCode int) {
	s.StatusCode = statusCode
	s.ResponseWriter.WriteHeader(statusCode)
}
