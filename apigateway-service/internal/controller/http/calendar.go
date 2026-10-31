package http

import (
	"net/http"

	"apigateway-service/pkg/clients"
	"github.com/go-chi/chi"
)

type CalendarController struct {
	UserServiceURL string
	HTTPClient     *clients.HTTPClient
}

func NewCalendarController(url string, cli *clients.HTTPClient) *CalendarController {
	return &CalendarController{UserServiceURL: url, HTTPClient: cli}
}

func (c *CalendarController) GetEventsHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/calendar/events")
}

func (c *CalendarController) CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/calendar/events")
}

func (c *CalendarController) DeleteEventHandler(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "event_id")
	proxyRequest(w, r, c.UserServiceURL, c.HTTPClient, "/api/v1/calendar/events/"+eventID)
}
