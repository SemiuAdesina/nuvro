package itinerary

import "time"

const AlgorithmVersion = "itinerary-v1"

type Pace string

const (
	Relaxed  Pace = "relaxed"
	Balanced Pace = "balanced"
	Packed   Pace = "packed"
)

type Shape struct {
	ActivitiesPerDay int
	BufferMinutes    int
	DayStartMinutes  int
	DayEndMinutes    int
	MealBreak        bool
}

func ShapeFor(pace Pace) Shape {
	switch pace {
	case Relaxed:
		return Shape{ActivitiesPerDay: 2, BufferMinutes: 45, DayStartMinutes: 10 * 60, DayEndMinutes: 19 * 60, MealBreak: true}
	case Packed:
		return Shape{ActivitiesPerDay: 4, BufferMinutes: 15, DayStartMinutes: 8 * 60, DayEndMinutes: 21 * 60, MealBreak: true}
	default:
		return Shape{ActivitiesPerDay: 3, BufferMinutes: 30, DayStartMinutes: 9 * 60, DayEndMinutes: 20 * 60, MealBreak: true}
	}
}

func ParsePace(raw string) Pace {
	switch Pace(raw) {
	case Relaxed:
		return Relaxed
	case Packed:
		return Packed
	default:
		return Balanced
	}
}

func (s Shape) Buffer() time.Duration { return time.Duration(s.BufferMinutes) * time.Minute }
