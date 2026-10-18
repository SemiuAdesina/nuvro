package change

import "errors"

var ErrNotClearable = errors.New("change: field cannot be cleared")

type state uint8

const (
	unchanged state = iota
	cleared
	assigned
)

type Change[T any] struct {
	state state
	value T
}

func Unchanged[T any]() Change[T] { return Change[T]{} }

func Set[T any](value T) Change[T] { return Change[T]{state: assigned, value: value} }

func Clear[T any]() Change[T] { return Change[T]{state: cleared} }

func (c Change[T]) IsUnchanged() bool { return c.state == unchanged }

func (c Change[T]) IsCleared() bool { return c.state == cleared }

func (c Change[T]) Value() (T, bool) { return c.value, c.state == assigned }

func (c Change[T]) ApplyTo(target *T) error {
	switch c.state {
	case unchanged:
		return nil
	case cleared:
		return ErrNotClearable
	default:
		*target = c.value
		return nil
	}
}

func (c Change[T]) ApplyToOptional(target **T) {
	switch c.state {
	case unchanged:
	case cleared:
		*target = nil
	default:
		value := c.value
		*target = &value
	}
}
