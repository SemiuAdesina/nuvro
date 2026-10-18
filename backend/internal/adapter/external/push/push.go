package push

import (
	"context"
	"errors"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
)

var ErrUnavailable = errors.New("push: no delivery provider configured")

type Platform string

const (
	Apple  Platform = "apple"
	Google Platform = "google"
)

type Device struct {
	UserID   string
	Platform Platform
	Token    string
}

type Notification struct {
	TripID     string
	AlertID    string
	TitleCode  string
	Severity   alert.Severity
	Action     alert.Action
	CollapseID string
}

type Sender interface {
	Send(context.Context, Device, Notification) error
}

type Registry interface {
	Devices(context.Context, string) ([]Device, error)
}

type Dispatcher struct {
	sender   Sender
	registry Registry
}

func NewDispatcher(sender Sender, registry Registry) *Dispatcher {
	return &Dispatcher{sender: sender, registry: registry}
}

func (d *Dispatcher) Configured() bool { return d.sender != nil && d.registry != nil }

func (d *Dispatcher) Dispatch(ctx context.Context, userID string, value alert.Alert) error {
	if !d.Configured() {
		return ErrUnavailable
	}
	devices, err := d.registry.Devices(ctx, userID)
	if err != nil {
		return err
	}
	notification := Notification{
		TripID: value.TripID, AlertID: value.ID, TitleCode: value.TitleCode,
		Severity: value.Severity, Action: value.Action, CollapseID: value.Fingerprint,
	}
	for _, device := range devices {
		if err := d.sender.Send(ctx, device, notification); err != nil {
			return err
		}
	}
	return nil
}
