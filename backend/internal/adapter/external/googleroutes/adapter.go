package googleroutes

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/httpclient"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/routing"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const DefaultBaseURL = "https://routes.googleapis.com"
const Provider = "google_routes"

const routeFieldMask = "routes.duration,routes.distanceMeters"
const matrixFieldMask = "originIndex,destinationIndex,duration,distanceMeters,condition"

type Adapter struct {
	client *httpclient.Client
	apiKey string
	timer  clock.Clock
}

func New(client *httpclient.Client, apiKey string, timer clock.Clock) *Adapter {
	return &Adapter{client: client, apiKey: apiKey, timer: timer}
}

func (a *Adapter) Compute(ctx context.Context, query routing.Query) (route.Evidence, error) {
	body := map[string]any{
		"origin":      waypoint(query.Origin),
		"destination": waypoint(query.Destination),
		"travelMode":  travelMode(query.Mode),
	}
	if query.Mode == route.Drive && !query.DepartAt.IsZero() {
		body["departureTime"] = query.DepartAt.UTC().Format(time.RFC3339)
		body["routingPreference"] = "TRAFFIC_AWARE"
	}
	var response computeResponse
	err := a.client.Post(ctx, httpclient.Request{
		Path: "/directions/v2:computeRoutes", Body: body, Headers: a.headers(routeFieldMask),
	}, &response)
	if err != nil {
		return route.Evidence{}, err
	}
	if len(response.Routes) == 0 {
		notFound := route.UnavailableEvidence(query.Origin, query.Destination, query.Mode)
		notFound.Status = route.NotFound
		return notFound, nil
	}
	return a.evidence(response.Routes[0].DistanceMeters, response.Routes[0].Duration, query.Origin, query.Destination, query.Mode), nil
}

func (a *Adapter) Matrix(ctx context.Context, query routing.MatrixQuery) (route.Matrix, error) {
	waypoints := make([]map[string]any, 0, len(query.Points))
	for _, point := range query.Points {
		waypoints = append(waypoints, map[string]any{"waypoint": waypoint(point)})
	}
	body := map[string]any{"origins": waypoints, "destinations": waypoints, "travelMode": travelMode(query.Mode)}
	var elements []matrixElement
	err := a.client.Post(ctx, httpclient.Request{
		Path: "/distanceMatrix/v2:computeRouteMatrix", Body: body, Headers: a.headers(matrixFieldMask),
	}, &elements)
	if err != nil {
		return route.Matrix{}, err
	}
	matrix := route.Matrix{Points: query.Points, Mode: query.Mode}
	for _, element := range elements {
		if !element.usable(len(query.Points)) {
			continue
		}
		matrix.Legs = append(matrix.Legs, route.Leg{
			FromIndex: element.OriginIndex, ToIndex: element.DestinationIndex,
			Evidence: a.evidence(element.DistanceMeters, element.Duration,
				query.Points[element.OriginIndex], query.Points[element.DestinationIndex], query.Mode),
		})
	}
	return matrix, nil
}

func (a *Adapter) evidence(distance int, duration string, origin, destination geo.Point, mode route.Mode) route.Evidence {
	seconds, ok := parseDuration(duration)
	if !ok || distance < 0 {
		return route.UnavailableEvidence(origin, destination, mode)
	}
	observedAt := a.timer.Now().UTC()
	return route.Evidence{
		Status: route.Found, Origin: origin, Destination: destination, Mode: mode,
		DistanceMeters: distance, DurationSeconds: seconds, ComputedAt: observedAt,
		Confidence: route.High,
		Source: evidence.Source{Provider: Provider, Freshness: evidence.Freshness{
			ObservedAt: observedAt, ValidUntil: observedAt.Add(routing.TimeToLive(mode)), Status: evidence.Live,
		}},
	}
}

func (a *Adapter) headers(fieldMask string) map[string]string {
	return map[string]string{"X-Goog-Api-Key": a.apiKey, "X-Goog-FieldMask": fieldMask}
}

func waypoint(point geo.Point) map[string]any {
	return map[string]any{"location": map[string]any{
		"latLng": map[string]any{"latitude": point.Latitude, "longitude": point.Longitude},
	}}
}

func travelMode(mode route.Mode) string {
	switch mode {
	case route.Walk:
		return "WALK"
	case route.Bike:
		return "BICYCLE"
	case route.Transit:
		return "TRANSIT"
	default:
		return "DRIVE"
	}
}

func parseDuration(raw string) (int, bool) {
	trimmed := strings.TrimSuffix(strings.TrimSpace(raw), "s")
	if trimmed == "" || trimmed != raw[:len(raw)-1] {
		return 0, false
	}
	seconds, err := strconv.Atoi(trimmed)
	if err != nil || seconds < 0 {
		return 0, false
	}
	return seconds, true
}

var _ routing.Provider = (*Adapter)(nil)
