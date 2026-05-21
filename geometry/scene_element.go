package geometry

import (
	"crypto/rand"
	"fmt"
)

type SceneElement struct {
	UUID string `json:"uuid"`
}

func NewSceneElement(uuid string) SceneElement {
	if uuid == "" {
		uuid = generateUUID()
	}

	return SceneElement{
		UUID: uuid,
	}
}

func (s *SceneElement) UUIDString() string {
	if s.UUID == "" {
		s.UUID = generateUUID()
	}
	return s.UUID
}

func generateUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("failed to generate UUID: %v", err))
	}

	// RFC 4122 variant and version bits.
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		b[0], b[1], b[2], b[3],
		b[4], b[5],
		b[6], b[7],
		b[8], b[9],
		b[10], b[11], b[12], b[13], b[14], b[15],
	)
}
