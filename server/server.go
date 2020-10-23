package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/OJOMB/url-analyser/config"
	"github.com/OJOMB/url-analyser/htmlanalyser"
	"github.com/gorilla/mux"
)

// AnalyseURLRequest is the expected structure of a POST body to /analyseUrl
type AnalyseURLRequest struct {
	URL string
}

// Server encapsulates the shared deps
type Server struct {
	router *mux.Router
	logger *log.Logger
	config *config.Config
}

// New returns a new Server instance
func New(mux *mux.Router, logger *log.Logger, config *config.Config) *Server {
	var s *Server = &Server{
		config: config,
		router: mux,
		logger: logger,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	fs := http.FileServer(http.Dir("./public"))
	handlePublic := http.StripPrefix("/public/", fs)
	s.router.PathPrefix("/public/").Handler(handlePublic)

	s.router.HandleFunc("/", s.handleIndex())
	s.router.HandleFunc("/analyseUrl", s.handleAnalyseURL()).Methods("POST")
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// ListenAndServe Listens and serves on the address specified in the server config
func (s *Server) ListenAndServe() {
	addr := fmt.Sprintf("%s:%d", s.config.IP, s.config.Port)
	s.logger.Printf("Server listening on: %s", addr)
	s.logger.Fatal(http.ListenAndServe(addr, s))
}

func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/app.html")
		fmt.Println("/ served")
	}
}

func (s *Server) handleAnalyseURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("analyseUrl endpoint hit")
		// read URL from POST request
		reqBody, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			s.logger.Print("Received unreadable request from client: " + err.Error())
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		var analyseURLRequest AnalyseURLRequest
		err = json.Unmarshal(reqBody, &analyseURLRequest)
		if err != nil {
			http.Error(w, "Received malformed JSON from client", http.StatusBadRequest)
			return
		}
		userInputtedURL := analyseURLRequest.URL
		s.logger.Printf("Received URL from client: %s", userInputtedURL)

		u, err := url.Parse(userInputtedURL)
		if err != nil {
			http.Error(w, "Received unparseable URL from client", http.StatusBadRequest)
			return
		}

		// constructed request to user submitted URL
		resp, err := http.Get(u.String())
		if err != nil {
			s.logger.Printf("Failed GET request to URL: %s. Got error: %s", u.String(), err.Error())
			http.Error(w, "Failed to retrieve data from request to url: "+u.String(), http.StatusInternalServerError)
			return
		}
		s.logger.Printf("Received response from: %s", u.String())

		respBody, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err == io.EOF {
			s.logger.Printf("Received empty response from request to URL: %s", u.String())
			s.logger.Print("Forwarding empty response to client")
			fmt.Fprintf(w, "")
			return
		} else if err != nil {
			s.logger.Printf("Received unreadable response from URL: %s. Encountered error: %s", u.String(), err.Error())
			http.Error(w, "Received unreadable response from URL: "+u.String(), http.StatusInternalServerError)
			return
		}
		document := string(respBody)
		s.logger.Print(document)
		analyser := htmlanalyser.New(document, u, s.logger)
		err = analyser.Analyse()
		if err != nil {
			http.Error(w, "Received unparseable HTML from URL: "+u.String(), http.StatusInternalServerError)
			return
		}

		responseDataForClient, err := json.Marshal(analyser)
		if err != nil {
			s.logger.Printf("Failed to marshal URL analysis to valid JSON response: %s", err.Error())
			http.Error(w, "Server Failed to construct a valid response", http.StatusInternalServerError)
			return
		}
		s.logger.Print(string(responseDataForClient))

		fmt.Fprintf(w, string(responseDataForClient))
	}
}
