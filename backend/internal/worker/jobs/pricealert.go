package jobs

import (
	"strconv"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/pricewatch"
)

const priceAlertLifetime = 48 * time.Hour

func priceAlert(watch pricewatch.Watch, change pricewatch.Change) alert.Alert {
	kind := alertTypeFor(change.Direction)
	value := alert.Alert{
		TripID: watch.TripID, Type: kind, Severity: severityFor(kind),
		TitleCode: "alert." + string(kind), RelatedItemID: watch.ID, Action: alert.ViewOffer,
		Status: alert.Unread, CreatedAt: change.ObservedAt, EffectiveAt: change.ObservedAt,
		ExpiresAt: change.ObservedAt.Add(priceAlertLifetime),
		MessageData: map[string]string{
			"product":    string(watch.Product),
			"oldAmount":  strconv.FormatInt(change.Baseline.Minor(), 10),
			"newAmount":  strconv.FormatInt(change.Current.Minor(), 10),
			"currency":   change.Baseline.Currency().String(),
			"percentage": strconv.Itoa(change.PercentagePoints),
		},
	}
	value.Fingerprint = alert.Fingerprint(watch.TripID, kind, watch.ID, priceWindow(change))
	return value
}

func alertTypeFor(direction pricewatch.Direction) alert.Type {
	switch direction {
	case pricewatch.Dropped:
		return alert.PriceDropped
	case pricewatch.Increased:
		return alert.PriceIncreased
	default:
		return alert.AvailabilityChanged
	}
}

func severityFor(kind alert.Type) alert.Severity {
	if kind == alert.PriceDropped {
		return alert.Attention
	}
	return alert.Info
}

func priceWindow(change pricewatch.Change) string {
	return change.ObservedAt.UTC().Format("2006-01-02") + ":" +
		strconv.FormatInt(change.Current.Minor(), 10)
}
