package configfile

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"golift.io/version"
)

// Template is the config file template.
//nolint: gochecknoglobals
var Template = template.Must(template.New("config").Funcs(Funcs()).Parse(tmpl))

func Funcs() template.FuncMap {
	return map[string]interface{}{
		"render": func(v interface{}) string {
			buf := new(bytes.Buffer)
			if err := toml.NewEncoder(buf).Encode(v); err != nil {
				return ""
			}
			return buf.String()
		},
		"version": func() string {
			return fmt.Sprintf("v%-7s @ %s", version.Version, time.Now().Format("060201T15:04"))
		},
	}
}

const tmpl = `################################################
# Notifiarr Client Example Configuration File. #
# Created by Notifiarr {{version}} #
################################################

# This API key must be copied from your notifiarr.com account.
api_key = "{{.APIKey}}"

## The ip:port or port to listen on for incoming web requests.
## You can use 127.0.0.1:5454 to listen only on localhost.
## This will be used in Plex for example to send webhooks, Example: http://localhost:5454
## This will be used in Media Requests for example to send payloads, Example: http://your-domain.com:5454
#
bind_addr = "{{.BindAddr}}"

## Quiet makes the app not log anything to output.
## Recommend setting log files if you make the app quiet.
## This is always true on Windows and macOS app.
## Log files are automatically written on those platforms.
#
quiet = {{.Quiet}}{{if .Debug}}
debug = true{{end}}

## All API paths start with /api. This does not affect incoming /plex webhooks.
## Change it to /somethingelse/api by setting urlbase to "/somethingelse"
#
urlbase = "{{.URLBase}}"

## Allowed upstream networks. The networks here are allowed to send x-forwarded-for.
## Set this to your reverse proxy server's IP or network. If you leave off the mask,
## then /32 or /128 is assumed depending on IP version. Empty by default. Example:
#
{{ if .Upstreams }}upstreams = {{render .Upstreams}}
{{ else }}#upstreams = [ "127.0.0.1/32", "::1/128" ]{{end}}

## If you provide a cert and key file (pem) paths, this app will listen with SSL/TLS.
## Uncomment both lines and add valid file paths. Make sure this app can read them.
#
{{if .SSLKeyFile}}ssl_key_file  = "{{.SSLKeyFile}}"{{else}}#ssl_key_file  = "/path/to/cert.key"{{end}}
{{if .SSLCrtFile}}ssl_cert_file  = "{{.SSLCrtFile}}"{{else}}#ssl_cert_file  = "/path/to/cert.key"{{end}}

## If you set these, logs will be written to these files.
## If blank on windows or macOS, log file paths are chosen for you.
{{if .LogFile}}log_file  = "{{.LogFile}}"{{else}}#log_file  = "~/.notifiarr/notifiarr.log"{{end}}
{{if .HTTPLog}}http_log  = "{{.HTTPLog}}"{{else}}#http_log  = "~/.notifiarr/notifiarr.http.log"{{end}}
#
## Set this to the number of megabytes to rotate files.
log_file_mb = {{.LogFileMb}}
#
## How many files to keep? 0 = all.
log_files = {{.LogFiles}}

## Web server and application timeouts.
#
timeout = "{{.Timeout}}"


##################
# Starr Settings #
##################

## The API keys are specific to the app. Get it from Settings -> General.
## Configurations for unused apps are harmless. Set URL and API key for
## apps you have and want to make requests to using Media Bot.
## See the Service Checks section below for information about setting the names.
##
## Examples follow.

{{if .Radarr}}{{range .Radarr}}[[radarr]]
  name     = "{{.Name}}"
  url      = "{{.URL}}"
  api_key  = "{{.APIKey}}"
  interval = "{{.Interval}}" # service check duration (if name is not empty)
  timeout  = "{{.Timeout}}"{{end -}}
{{else}}#[[radarr]]
#name    = "" # set a name to enable checks of your service.
#url     = "http://127.0.0.1:7878/radarr"
#api_key = ""{{end}}

{{if .Readarr}}{{range .Readarr}}[[readarr]]
  name     = "{{.Name}}"
  url      = "{{.URL}}"
  api_key  = "{{.APIKey}}"
  interval = "{{.Interval}}" # service check duration (if name is not empty)
  timeout  = "{{.Timeout}}"{{end -}}
{{else}}#[[readarr]]
#name    = "" # set a name to enable checks of your service.
#url     = "http://127.0.0.1:8787/readarr"
#api_key = ""{{end}}

{{if .Sonarr}}{{range .Sonarr}}[[sonarr]]
  name     = "{{.Name}}"
  url      = "{{.URL}}"
  api_key  = "{{.APIKey}}"
  interval = "{{.Interval}}" # service check duration (if name is not empty)
  timeout  = "{{.Timeout}}"{{end -}}
{{else}}#[[sonarr]]
#name    = "" # set a name to enable checks of your service.
#url     = "http://sonarr:8989/"
#api_key = ""{{end}}

{{if .Lidarr}}{{range .Lidarr}}[[lidarr]]
  name     = "{{.Name}}"
  url      = "{{.URL}}"
  api_key  = "{{.APIKey}}"
  interval = "{{.Interval}}" # service check duration (if name is not empty)
  timeout  = "{{.Timeout}}"{{end -}}
{{else}}#[[lidarr]]
#name    = "" # set a name to enable checks of your service.
#url     = "http://lidarr:8989/"
#api_key = ""{{end}}


#################
# Plex Settings #
#################

## Find your token: https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token/
#
[plex]
{{if .Plex}}url         = "{{.Plex.URL}}" # Your plex URL
token       = "{{.Plex.Token}}"     # your plex token; get this from a web inspector
interval    = "{{.Plex.Interval}}"  # how often to send session data, 0 = off
cooldown    = "{{.Plex.Cooldown}}"  # how often plex webhooks may trigger session hooks
account_map = "{{.Plex.AccountMap}}"     # map an email to a name, ex: "som@ema.il,Name|some@ther.mail,name"
server      = "{{.Plex.Server}}" # optional name of the server the notifications are from
movies_percent_complete = {{.Plex.MoviesPC}} # 0, 70-99, send notifications when a movie session is this % complete.
series_percent_complete = {{.Plex.SeriesPC}} # 0, 70-99, send notifications when an episode session is this % complete.
{{else -}}
#url         = "http://localhost:32400" # Your plex URL
#token       = ""     # your plex token; get this from a web inspector
#interval    = "30m"  # how often to send session data, 0 = off
#cooldown    = "15s"  # how often plex webhooks may trigger session hooks
#account_map = ""     # shared plex servers: map an email to a name, ex: "som@ema.il,Name|some@ther.mail,name"
#server      = "plex" # optional name of the server the notifications are from
#movies_percent_complete = 0 # 0, 70-99, send notifications when a movie session is this % complete.
#series_percent_complete = 0 # 0, 70-99, send notifications when an episode session is this % complete.
{{- end }}

#####################
# Snapshot Settings #
#####################

## Install package(s)
##	- Windows:  smartmontools - https://sourceforge.net/projects/smartmontools/
##	- Linux:    apt install smartmontools || yum install smartmontools
##	- Synology: opkg install smartmontools
##	- Entware:  https://github.com/Entware/Entware-ng/wiki/Install-on-Synology-NAS
##  - Entware Package List:  https://github.com/Entware/Entware-ng/wiki/Install-on-Synology-NAS
#
[snapshot]
interval          = "{{.Snapshot.Interval}}" # how often to send a snapshot, 0 = off
timeout           = "{{.Snapshot.Timeout}}" # how long a snapshot may take
monitor_raid      = {{.Snapshot.Raid}}  # mdadm / megacli
monitor_drives    = {{.Snapshot.DriveData}}  # smartctl: age, temp, health
monitor_space     = {{.Snapshot.DiskUsage}}  # disk usage for all partitions
monitor_uptime    = {{.Snapshot.Uptime}}  # system data, users, hostname, uptime, os, build
monitor_cpuMemory = {{.Snapshot.CPUMem}}  # literally cpu usage, load averages, and memory
monitor_cpuTemp   = {{.Snapshot.CPUTemp}}  # cpu temperatures, not available on all platforms
{{- if .Snapshot.ZFSPools}}
zfs_pools         = {{render .Snapshot.ZFSPools}}    # list of zfs pools, ex: zfs_pools=["data", "data2"]
{{else}}
zfs_pools         = []   # list of zfs pools, ex: zfs_pools=["data", "data2"]
{{end -}}
use_sudo          = {{.Snapshot.UseSudo}} # sudo is needed on unix when monitor_drives=true or for megacli.
# Example sudoers entries follow. Fix the paths to smartctl and MegaCli.
# notifiarr ALL=(root) NOPASSWD:/usr/sbin/smartctl *
# notifiarr ALL=(root) NOPASSWD:/usr/sbin/MegaCli64 -LDInfo -Lall -aALL

##################
# Service Checks #
##################

## This application performs service checks on configured services at the specified interval.
## The service states are sent to Notifiarr.com. Failed services generate a notification.
## Setting names on Starr apps (above) enables service checks for that app.
## Use the [[service]] directive to add more service checks. Example below.

[services]
  disabled = {{.Services.Disabled}}   # Setting this to true disables all service checking routines.
  interval = "{{.Services.Interval}}" # How often to send service states to Notifiarr.com. Minimum = 5m.
  parallel = {{.Services.Parallel}}       # How many services to check concurrently. 1 should be enough.

## Uncomment the following section to create a service check on a URL or IP:port.
## You may include as many [[service]] sections as you have services to check.
## Do not add Radarr, Sonarr, Readarr or Lidarr here! Add a name to enable their checks.
#
{{if .Service}}{{range .Service}}[[service]]
  name     = "{{.Name}}"
  type     = "{{.Type}}"
  check    = "{{.Check}}"
  expect   = "{{.Expect}}"
  timeout  = "{{.Timeout}}"
  interval = "{{.Interval}}"{{end}}
{{else}}# Example with comments follows.
#{{end}}
#[[service]]
#  name     = "MyServer"          # name must be unique
#  type     = "http"              # type can be "http" or "tcp"
#  check    = "http://127.0.0.1"  # url for 'http', host/IP:port for 'tcp'
#  expect   = "200"               # return code to expect (for http only)
#  timeout  = "10s"               # how long to wait for tcp or http checks.
#  interval = "5m"                # how often to check this service.

{{if not .Service}}## Another example. Remember to uncomment [[service]] if you use this!
#
#[[service]]
#  name    = "Bazarr"
#  type    = "http"
#  check   = "http://10.1.1.2:6767/series/"
#  expect  = "200"
#  timeout = "10s"{{end}}
`