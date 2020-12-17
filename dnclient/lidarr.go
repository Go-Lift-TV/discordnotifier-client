package dnclient

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"golift.io/starr"
	"golift.io/starr/lidarr"
)

/*
[9:19 PM] nitsua: mbid i think is lidarr
[9:19 PM] nitsua: music brainz i believe is the source for it
*/

func (c *Config) fixLidarrConfig() {
	for i := range c.Lidarr {
		if c.Lidarr[i].Timeout.Duration == 0 {
			c.Lidarr[i].Timeout.Duration = c.Timeout.Duration
		}
	}
}

// LidarrConfig represents the input data for a Lidarr server.
type LidarrConfig struct {
	*starr.Config
	*lidarr.Lidarr
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

func (c *Client) logLidarr() {
	if count := len(c.Lidarr); count == 1 {
		c.Printf(" => Lidarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Lidarr[0].URL, c.Lidarr[0].APIKey != "", c.Lidarr[0].Timeout, c.Lidarr[0].ValidSSL)
	} else {
		c.Print(" => Lidarr Config:", count, "servers")

		for _, f := range c.Lidarr {
			c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
}

// getLidarr finds a Lidarr based on the passed-in ID.
// Every Lidarr handler calls this.
func (c *Client) getLidarr(id string) *LidarrConfig {
	j, _ := strconv.Atoi(id)

	for i, app := range c.Lidarr {
		if i != j-1 { // discordnotifier wants 1-indexes
			continue
		}

		return app
	}

	return nil
}

func (c *Client) lidarrRootFolders(r *http.Request) (int, interface{}) {
	// Make sure the provided lidarr id exists.
	lidar := c.getLidarr(mux.Vars(r)["id"])
	if lidar == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", mux.Vars(r)["id"], ErrNoLidarr)
	}

	// Get folder list from Lidarr.
	folders, err := lidar.GetRootFolders()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting folders: %w", err)
	}

	// Format folder list into a nice path=>freesSpace map.
	p := make(map[string]int64)
	for i := range folders {
		p[folders[i].Path] = folders[i].FreeSpace
	}

	return http.StatusOK, p
}

func (c *Client) lidarrProfiles(r *http.Request) (int, interface{}) {
	// Make sure the provided lidarr id exists.
	lidar := c.getLidarr(mux.Vars(r)["id"])
	if lidar == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", mux.Vars(r)["id"], ErrNoLidarr)
	}

	// Get the profiles from lidarr.
	profiles, err := lidar.GetQualityProfiles()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	// Format profile ID=>Name into a nice map.
	p := make(map[int]string)
	for i := range profiles {
		p[profiles[i].ID] = profiles[i].Name
	}

	return http.StatusOK, p
}

func (c *Client) lidarrQualityDefs(r *http.Request) (int, interface{}) {
	// Make sure the provided lidarr id exists.
	lidar := c.getLidarr(mux.Vars(r)["id"])
	if lidar == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", mux.Vars(r)["id"], ErrNoLidarr)
	}

	// Get the profiles from lidarr.
	definitions, err := lidar.GetQualityDefinition()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	// Format definitions ID=>Title into a nice map.
	p := make(map[int]string)
	for i := range definitions {
		p[definitions[i].ID] = definitions[i].Title
	}

	return http.StatusOK, p
}

func (c *Client) lidarrAddAlbum(r *http.Request) (int, interface{}) {
	return http.StatusLocked, "lidarr does not work yet"
}
