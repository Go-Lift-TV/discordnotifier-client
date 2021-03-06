package client

/*
  This file contains the procedures that validate config data and initialize each app.
  All startup logs come from below. Every procedure in this file is run once on startup.
*/

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/ui"
)

const disabled = "disabled"

// PrintStartupInfo prints info about our startup config.
// This runs once on startup, and again during reloads.
func (c *Client) PrintStartupInfo() {
	c.Printf("==> %s <==", mnd.HelpLink)
	c.Print("==> Startup Settings <==")
	c.printSonarr()
	c.printRadarr()
	c.printLidarr()
	c.printReadarr()
	c.printPlex()
	c.Printf(" => Timeout: %v, Quiet: %v", c.Config.Timeout, c.Config.Quiet)
	c.Printf(" => Trusted Upstream Networks: %v", c.Config.Allow)

	if c.Config.SSLCrtFile != "" && c.Config.SSLKeyFile != "" {
		c.Print(" => Web HTTPS Listen:", "https://"+c.Config.BindAddr+path.Join("/", c.Config.URLBase))
		c.Print(" => Web Cert & Key Files:", c.Config.SSLCrtFile+", "+c.Config.SSLKeyFile)
	} else {
		c.Print(" => Web HTTP Listen:", "http://"+c.Config.BindAddr+path.Join("/", c.Config.URLBase))
	}

	c.printLogFileInfo()
}

func (c *Client) printLogFileInfo() {
	if c.Config.LogFile != "" {
		if c.Config.LogFiles > 0 {
			c.Printf(" => Log File: %s (%d @ %dMb)", c.Config.LogFile, c.Config.LogFiles, c.Config.LogFileMb)
		} else {
			c.Printf(" => Log File: %s (no rotation)", c.Config.LogFile)
		}
	}

	if c.Config.HTTPLog != "" {
		if c.Config.LogFiles > 0 {
			c.Printf(" => HTTP Log: %s (%d @ %dMb)", c.Config.HTTPLog, c.Config.LogFiles, c.Config.LogFileMb)
		} else {
			c.Printf(" => HTTP Log: %s (no rotation)", c.Config.HTTPLog)
		}
	}

	if c.Config.Debug && c.Config.DebugLog != "" {
		if c.Config.LogFiles > 0 {
			c.Printf(" => Debug Log: %s (%d @ %dMb)", c.Config.DebugLog, c.Config.LogFiles, c.Config.LogFileMb)
		} else {
			c.Printf(" => Debug Log: %s (no rotation)", c.Config.DebugLog)
		}
	}

	if c.Config.Services.LogFile != "" && !c.Config.Services.Disabled && len(c.Config.Service) > 0 {
		if c.Config.LogFiles > 0 {
			c.Printf(" => Service Checks Log: %s (%d @ %dMb)", c.Config.Services.LogFile, c.Config.LogFiles, c.Config.LogFileMb)
		} else {
			c.Printf(" => Service Checks Log: %s (no rotation)", c.Config.Services.LogFile)
		}
	}
}

// Exit stops the web server and logs our exit messages. Start() calls this.
func (c *Client) Exit() (err error) {
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), c.Config.Timeout.Duration)
		defer cancel()

		if c.server != nil {
			err = c.server.Shutdown(ctx)
		}
	}()

	if c.sigkil == nil || c.sighup == nil {
		return
	}

	for {
		select {
		case sigc := <-c.sigkil:
			c.Printf("[%s] Need help? %s\n=====> Exiting! Caught Signal: %v", c.Flags.Name(), mnd.HelpLink, sigc)
			return
		case sigc := <-c.sighup:
			c.checkReloadSignal(sigc)
		}
	}
}

// reloadConfiguration is called from a menu tray item or when a HUP signal is received.
// Re-reads the configuration file and stops/starts all the internal routines.
// Also closes and re-opens all log files. Any errors cause the application to exit.
func (c *Client) reloadConfiguration(msg string) {
	c.Print("==> Reloading Configuration: " + msg)
	c.notify.Stop()
	c.Config.Services.Stop()

	if err := c.StopWebServer(); err != nil && !errors.Is(err, ErrNoServer) {
		c.Errorf("Reloading Config (1): %v\nNotifiarr EXITING!", err)
		panic(err)
	} else if !errors.Is(err, ErrNoServer) {
		defer c.StartWebServer()
	}

	if err := c.Config.Get(c.Flags.ConfigFile, c.Flags.EnvPrefix); err != nil {
		c.Errorf("Reloading Config (2): %v\nNotifiarr EXITING!", err)
		panic(err)
	}

	if errs := c.Logger.Close(); len(errs) > 0 {
		// in a go routine in case logging is blocked
		go c.Errorf("Reloading Config (3): %v\nNotifiarr EXITING!", errs)
		time.Sleep(1 * time.Second)
		panic(errs)
	}

	c.Logger.SetupLogging(c.Config.LogConfig)
	c.configureServices(true)

	if err := c.Config.Services.Start(c.Config.Service); err != nil {
		c.Errorf("Reloading Config (4): %v\nNotifiarr EXITING!", err)
		panic(fmt.Errorf("service checks: %w", err))
	}

	c.Print("==> Configuration Reloaded! Config File:", c.Flags.ConfigFile)
	_, _ = ui.Info(mnd.Title, "Configuration Reloaded!") //nolint:wsl
}

// printPlex is called on startup to print info about configured Plex instance(s).
func (c *Client) printPlex() {
	p := c.Config.Plex
	if !p.Configured() {
		return
	}

	name := p.Name
	if name == "" {
		name = "<possible connection error>"
	}

	c.Printf(" => Plex Config: 1 server: %s @ %s (enables incoming APIs and webhook)", name, p.URL)
}

// printLidarr is called on startup to print info about each configured server.
func (c *Client) printLidarr() {
	if len(c.Config.Lidarr) == 1 {
		f := c.Config.Lidarr[0]

		checkQ := disabled
		if f.CheckQ != nil {
			checkQ = strconv.Itoa(int(*f.CheckQ))
		}

		c.Printf(" => Lidarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v, check_q: %s",
			f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, checkQ)

		return
	}

	c.Print(" => Lidarr Config:", len(c.Config.Lidarr), "servers")

	for _, f := range c.Config.Lidarr {
		checkQ := disabled
		if f.CheckQ != nil {
			checkQ = strconv.Itoa(int(*f.CheckQ))
		}

		c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v, check_q: %s",
			f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, checkQ)
	}
}

// printRadarr is called on startup to print info about each configured server.
func (c *Client) printRadarr() {
	if len(c.Config.Radarr) == 1 {
		f := c.Config.Radarr[0]

		checkQ := disabled
		if f.CheckQ != nil {
			checkQ = strconv.Itoa(int(*f.CheckQ))
		}

		c.Printf(" => Radarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v, check_q: %s",
			f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, checkQ)

		return
	}

	c.Print(" => Radarr Config:", len(c.Config.Lidarr), "servers")

	for _, f := range c.Config.Radarr {
		checkQ := disabled
		if f.CheckQ != nil {
			checkQ = strconv.Itoa(int(*f.CheckQ))
		}

		c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v, check_q: %s",
			f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, checkQ)
	}
}

// printReadarr is called on startup to print info about each configured server.
func (c *Client) printReadarr() {
	if len(c.Config.Readarr) == 1 {
		f := c.Config.Readarr[0]

		checkQ := disabled
		if f.CheckQ != nil {
			checkQ = strconv.Itoa(int(*f.CheckQ))
		}

		c.Printf(" => Readarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v, check_q: %s",
			f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, checkQ)

		return
	}

	c.Print(" => Readarr Config:", len(c.Config.Lidarr), "servers")

	for _, f := range c.Config.Readarr {
		checkQ := disabled
		if f.CheckQ != nil {
			checkQ = strconv.Itoa(int(*f.CheckQ))
		}

		c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v, check_q: %s",
			f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, checkQ)
	}
}

// printSonarr is called on startup to print info about each configured server.
func (c *Client) printSonarr() {
	if len(c.Config.Sonarr) == 1 {
		f := c.Config.Sonarr[0]

		checkQ := disabled
		if f.CheckQ != nil {
			checkQ = strconv.Itoa(int(*f.CheckQ))
		}

		c.Printf(" => Sonarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v, check_q: %s",
			f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, checkQ)

		return
	}

	c.Print(" => Sonarr Config:", len(c.Config.Lidarr), "servers")

	for _, f := range c.Config.Sonarr {
		checkQ := disabled
		if f.CheckQ != nil {
			checkQ = strconv.Itoa(int(*f.CheckQ))
		}

		c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v, check_q: %s",
			f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, checkQ)
	}
}
