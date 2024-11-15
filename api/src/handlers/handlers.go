package handlers

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"infraops.dev/statuspage-core/config"
	db_migrations "infraops.dev/statuspage-core/db/migrations"
	"infraops.dev/statuspage-core/utils"
)

type ServiceState struct {
	Host              string
	IsUp              bool
	PreviousIsUp      bool
	IsPreviousIsUpSet bool
	LastChange        time.Time
	UpdatetimeStart   time.Time
	LastRequestTime   time.Time
}

type ServiceStates struct {
	states map[string]*ServiceState
	mu     sync.Mutex
}

var serviceStates = ServiceStates{states: make(map[string]*ServiceState)}

type UpdatetimeEvent struct {
	Time   time.Time
	Reason string
}

func LogUpdatetimeEvent(event UpdatetimeEvent, createNotification bool) {
	db_migrations.InsertLogEntry(config.DbConn, db_migrations.LogEntry{
		Timestamp: event.Time.Local().Format(time.RFC3339),
		Level:     "INFO",
		Message:   event.Reason,
	})
}

func (ss *ServiceStates) UpdateServiceState(host string, isCurrentlyUp bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	currentState, exists := ss.states[host]
	if !exists {
		log.Printf("Added host '%s' to the list of monitored hosts", host)
		now := time.Now()
		ss.states[host] = &ServiceState{
			Host:            host,
			IsUp:            isCurrentlyUp,
			LastChange:      now,
			UpdatetimeStart: now,
			LastRequestTime: now,
		}
		return
	}

	currentState.LastRequestTime = time.Now()

	if currentState.IsUp != isCurrentlyUp {
		currentState.IsUp = isCurrentlyUp
		currentState.LastChange = time.Now()

		if !isCurrentlyUp {
			currentState.UpdatetimeStart = time.Now()
		} else {
			currentState.UpdatetimeStart = time.Time{}
		}

		status := "down"
		if isCurrentlyUp {
			status = "up"
		}
		LogUpdatetimeEvent(UpdatetimeEvent{
			Time:   time.Now(),
			Reason: fmt.Sprintf("aHost '%s' status changed to %s", host, status),
		}, true)
	}
}

var hosts = []string{}

func MonitorHostChanges(timeout time.Duration) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	logAndUpdate := func(reason string, currentTime time.Time) {
		log.Println(reason)
		LogUpdatetimeEvent(UpdatetimeEvent{
			Time:   currentTime,
			Reason: reason,
		}, true)
	}

	for range ticker.C {
		currentTime := time.Now()
		serviceStates.mu.Lock()
		for host, state := range serviceStates.states {
			updatedState := state

			if !updatedState.IsPreviousIsUpSet {
				updatedState.IsPreviousIsUpSet = true
				updatedState.PreviousIsUp = updatedState.IsUp
				reason := fmt.Sprintf("Initialized Host %s status and set it to %t", host, updatedState.IsUp)
				logAndUpdate(reason, currentTime)
				serviceStates.states[host] = updatedState
			} else if updatedState.IsUp != updatedState.PreviousIsUp {
				updatedState.PreviousIsUp = updatedState.IsUp
				reason := fmt.Sprintf("Updating Host %s status and changing it to %t", host, updatedState.IsUp)
				logAndUpdate(reason, currentTime)
				serviceStates.states[host] = updatedState
			} else {
				if !slices.Contains(hosts, host) {
					reason := fmt.Sprintf("Added Host %s to monitored hosts-list with an initialization-state of %t", host, updatedState.IsUp)
					logAndUpdate(reason, currentTime)
					hosts = append(hosts, host)
					serviceStates.states[host] = updatedState
				}
			}
		}
		serviceStates.mu.Unlock()
	}
}

func ensurePort(host string) string {
	if !strings.Contains(host, ":") {
		return host + ":443"
	}
	return host
}

func HandleUp(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Query().Get("host")
	if host == "" {
		utils.HttpError(w, "Host parameter is required", http.StatusBadRequest)
		return
	}

	hostWithPort := ensurePort(host)
	path := r.URL.Query().Get("path")
	metrics, err := utils.HostMetrics(hostWithPort, path)
	if err != nil {
		utils.HttpError(w, "Failed to metrics info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	serviceStates.UpdateServiceState(host, metrics.Reachable)

	responseData := map[string]interface{}{
		"reachable":         metrics.Reachable,
		"dnsResolutionTime": metrics.DnsResolutionTime.String(),
		"tcpConnectionTime": metrics.TcpConnectionTime.String(),
		"tlsConnectionTime": metrics.TlsConnectionTime.String(),
		"httpTime":          metrics.HttpTime.String(),
		"statusCode":        metrics.StatusCode,
		"error":             fmt.Errorf("%f", metrics.Error),
	}

	utils.JsonResponse(w, responseData)
}

func HandleCertInfo(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Query().Get("host")
	if host == "" {
		utils.HttpError(w, "Host parameter is required", http.StatusBadRequest)
		return
	}

	hostWithPort := ensurePort(host)
	metrics, err := utils.HostMetrics(hostWithPort, "")

	if err != nil {
		utils.HttpError(w, "Failed to fetch metrics info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !metrics.Reachable {
		utils.HttpError(w, "Host is not reachable", http.StatusServiceUnavailable)
		return
	}

	certInfo, err := utils.FetchCertInfo(hostWithPort)
	if err != nil {
		utils.HttpError(w, "Failed to fetch cert info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JsonResponse(w, certInfo)
}
