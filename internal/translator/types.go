package translator

import (
	"encoding/json"
	"math"
)

func specialPtrFloat[T ~float32 | ~float64](f *T) *SpecialFloat64 {
	if f == nil {
		return nil
	}

	v := SpecialFloat64(*f)
	return &v
}

type SpecialFloat64 float64

func (s *SpecialFloat64) UnmarshalJSON(data []byte) error {
	// log.Printf("UnmarshalJSON data: %s", data)
	var v float64
	if string(data) == "null" {
		// log.Printf("UnmarshalJSON null detected")
		*s = 0
		return nil
	}
	err := json.Unmarshal(data, &v)
	if err != nil {
		// log.Printf("UnmarshalJSON error: %v", err)
		return err
	}
	*s = SpecialFloat64(v)
	return nil
}

func (s *SpecialFloat64) MarshalJSON() ([]byte, error) {
	// log.Printf("MarshalJSON value: %v", s)
	if s == nil {
		// log.Printf("MarshalJSON nil detected")
		return []byte("null"), nil
	}
	if math.IsNaN(float64(*s)) {
		// log.Printf("MarshalJSON NaN detected")
		return []byte("null"), nil
	}
	// log.Printf("MarshalJSON normal value")
	// Convert to float64 and marshal using the standard encoding
	// This avoids recursion by not calling MarshalJSON on SpecialFloat64 again
	type plainFloat float64
	return json.Marshal(plainFloat(*s))
}
