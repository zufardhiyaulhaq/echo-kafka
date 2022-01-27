package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/evio"
	"github.com/zufardhiyaulhaq/echo-kafka/pkg/producer"
	"github.com/zufardhiyaulhaq/echo-kafka/pkg/settings"
)

type Server struct {
	settings settings.Settings
	client   producer.Producer
}

func NewServer(settings settings.Settings, client producer.Producer) Server {
	return Server{
		settings: settings,
		client:   client,
	}
}

func (e Server) ServeEcho() {
	var events evio.Events

	events.Data = func(c evio.Conn, in []byte) (out []byte, action evio.Action) {
		value := string(in)
		err := e.client.SendMessage(e.settings.KafkaTopic, value)

		if err != nil {
			out = []byte(err.Error())
			return
		}

		out = []byte("success produce to kafka")
		return
	}

	if err := evio.Serve(events, "tcp://0.0.0.0:"+e.settings.EchoPort); err != nil {
		log.Fatal().Err(err)
	}
}

func (e Server) ServeHTTP() {
	handler := NewHandler(e.settings, e.client)

	r := mux.NewRouter()

	r.HandleFunc("/kafka/{key}", handler.Handle)
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello!"))
	})
	r.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello!"))
	})

	err := http.ListenAndServe(":"+e.settings.HTTPPort, r)
	if err != nil {
		log.Fatal().Err(err)
	}
}

type Handler struct {
	settings settings.Settings
	client   producer.Producer
}

func NewHandler(settings settings.Settings, client producer.Producer) Handler {
	return Handler{
		settings: settings,
		client:   client,
	}
}

func (h Handler) Handle(w http.ResponseWriter, req *http.Request) {
	value := mux.Vars(req)["key"]

	if h.settings.KafkaEnableProducer {
		err := h.client.SendMessage(h.settings.KafkaTopic, value)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success produce to kafka"))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("producer is disabled"))
	}

}
