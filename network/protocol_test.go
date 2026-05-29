package network

import (
	"reflect"
	"testing"
)

func TestSerialization(t *testing.T) {
	msg := Message{
		Command: CmdShot,
		Args: map[string]string{
			"x": "1",
			"y": "2",
		},
	}

	data, err := SerializeMessage(msg)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	decoded, err := DeserializeMessage(data)
	if err != nil {
		t.Fatalf("failed to deserialize: %v", err)
	}

	if !reflect.DeepEqual(msg, *decoded) {
		t.Errorf("expected %+v, got %+v", msg, *decoded)
	}
}
