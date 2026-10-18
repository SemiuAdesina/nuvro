package itinerary

import (
	"slices"
	"time"
)

type Itinerary struct {
	ID               string
	TripID           string
	Timezone         string
	Version          int
	GeneratedAt      time.Time
	Items            []Item
	AlgorithmVersion string
}

type Day struct {
	Number int
	Date   time.Time
	Items  []Item
}

func (i Itinerary) Days() []Day {
	location := i.location()
	byDay := map[int][]Item{}
	for _, item := range i.Items {
		byDay[item.Day] = append(byDay[item.Day], item)
	}
	numbers := make([]int, 0, len(byDay))
	for number := range byDay {
		numbers = append(numbers, number)
	}
	slices.Sort(numbers)
	days := make([]Day, 0, len(numbers))
	for _, number := range numbers {
		items := byDay[number]
		slices.SortStableFunc(items, func(left, right Item) int { return left.StartAt.Compare(right.StartAt) })
		days = append(days, Day{Number: number, Date: items[0].StartAt.In(location), Items: items})
	}
	return days
}

func (i Itinerary) LockedItems() []Item {
	locked := make([]Item, 0, len(i.Items))
	for _, item := range i.Items {
		if item.Locked() {
			locked = append(locked, item)
		}
	}
	return locked
}

func (i Itinerary) location() *time.Location {
	if i.Timezone == "" {
		return time.UTC
	}
	location, err := time.LoadLocation(i.Timezone)
	if err != nil {
		return time.UTC
	}
	return location
}
