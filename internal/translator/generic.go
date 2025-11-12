package translator

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
)

type ProducesJSON interface {
	ToJSON() ([]byte, error)
}

type Meshtastic[T proto.Message, P ProducesJSON] struct {
	f func(T) P
}

func New[T proto.Message, P ProducesJSON](f func(T) P) *Meshtastic[T, P] {
	return &Meshtastic[T, P]{f: f}
}

func (m *Meshtastic[T, P]) Convert(in T) (P, error) {
	return m.f(in), nil
}

func (m *Meshtastic[T, P]) Decode(in []byte) (P, error) {
	// log.Printf("Convert called with %d bytes", len(in))
	var payload T
	var zero P
	var ok bool
	payload, ok = payload.ProtoReflect().New().Interface().(T)
	if !ok {
		return zero, errors.New("unable to create new instance of payload type")
	}
	if err := proto.Unmarshal(in, payload); err != nil {
		return zero, fmt.Errorf("%w: unable to unmarshal payload", err)
	}
	return m.f(payload), nil
}

func (m *Meshtastic[T, P]) UnmarshalAndConvert(in []byte) ([]byte, error) {
	// log.Printf("UnmarshalAndConvert called with %d bytes", len(in))
	src, err := m.Decode(in)
	if err != nil {
		return nil, err
	}

	return src.ToJSON()
}
