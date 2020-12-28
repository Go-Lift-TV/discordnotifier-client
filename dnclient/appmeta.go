package dnclient

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"golift.io/starr"
	"golift.io/starr/lidarr"
	"golift.io/starr/radarr"
	"golift.io/starr/readarr"
	"golift.io/starr/sonarr"
)

/* This file contains:
 *** The middleware procedure that stores the app interface in a request context.
 *** Procedures to save and fetch an app interface into/from a request content.
 */

// App allows safely storing context values.
type App string

// Constant for each app to unique identify itself.
// These strings are also used as a suffix to the /api/ web path.
const (
	Sonarr  App = "sonarr"
	Readarr App = "readarr"
	Radarr  App = "radarr"
	Lidarr  App = "lidarr"
)

// LidarrConfig represents the input data for a Lidarr server.
type LidarrConfig struct {
	*starr.Config
	*lidarr.Lidarr
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

// RadarrConfig represents the input data for a Radarr server.
type RadarrConfig struct {
	*starr.Config
	*radarr.Radarr
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

// ReadarrConfig represents the input data for a Readarr server.
type ReadarrConfig struct {
	*starr.Config
	*readarr.Readarr
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

// SonarrConfig represents the input data for a Sonarr server.
type SonarrConfig struct {
	*starr.Config
	*sonarr.Sonarr
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

// apiHandler is our custom handler function for APIs.
type apiHandler func(r *http.Request) (int, interface{})

// handleAPIpath makes adding API paths a little cleaner.
// This grabs the app struct and saves it in a context before calling the handler.
func (c *Client) handleAPIpath(app App, api string, next apiHandler, method ...string) {
	if len(method) == 0 {
		method = []string{"GET"}
	}

	id := "{id:[0-9]+}"
	if app == "" {
		id = ""
	}

	// disccordnotifier uses 1-indexes.
	c.router.Handle(path.Join("/", c.Config.URLBase, "api", string(app), id, api),
		c.checkAPIKey(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now() // Capture the starr app request time in a response header.
			switch id, _ := strconv.Atoi(mux.Vars(r)["id"]); {
			default: // unknown app, just run the handler.
				i, m := next(r)
				c.respond(w, i, m, start)
			case app == Radarr && (id > len(c.Config.Radarr) || id < 1):
				c.respond(w, http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", id, ErrNoRadarr), start)
			case app == Lidarr && (id > len(c.Config.Lidarr) || id < 1):
				c.respond(w, http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", id, ErrNoLidarr), start)
			case app == Sonarr && (id > len(c.Config.Sonarr) || id < 1):
				c.respond(w, http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", id, ErrNoSonarr), start)
			case app == Readarr && (id > len(c.Config.Readarr) || id < 1):
				c.respond(w, http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", id, ErrNoReadarr), start)

			// These store the application configuration (starr) in a context then pass that into the next method.
			// They retrieve the return code and output, then send a response (c.respond).
			case app == Radarr:
				i, m := next(r.WithContext(context.WithValue(r.Context(), Radarr, c.Config.Radarr[id-1])))
				c.respond(w, i, m, start)
			case app == Lidarr:
				i, m := next(r.WithContext(context.WithValue(r.Context(), Lidarr, c.Config.Lidarr[id-1])))
				c.respond(w, i, m, start)
			case app == Sonarr:
				i, m := next(r.WithContext(context.WithValue(r.Context(), Sonarr, c.Config.Sonarr[id-1])))
				c.respond(w, i, m, start)
			case app == Readarr:
				i, m := next(r.WithContext(context.WithValue(r.Context(), Readarr, c.Config.Readarr[id-1])))
				c.respond(w, i, m, start)
			}
		}))).Methods(method...)
}

// checkAPIKey drops a 403 if the API key doesn't match.
func (c *Client) checkAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != c.Config.APIKey {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

/* Every API call runs one of these methods to find the interface for the respective app. */

func getLidarr(r *http.Request) *LidarrConfig {
	return r.Context().Value(Lidarr).(*LidarrConfig)
}

func getRadarr(r *http.Request) *RadarrConfig {
	return r.Context().Value(Radarr).(*RadarrConfig)
}

func getReadarr(r *http.Request) *ReadarrConfig {
	return r.Context().Value(Readarr).(*ReadarrConfig)
}

func getSonarr(r *http.Request) *SonarrConfig {
	return r.Context().Value(Sonarr).(*SonarrConfig)
}
