package message

type Response struct {
	HeadLength uint32
	BodyLength uint32
	RequestID  uint32
	Version    uint32
	Comresser  uint32
	Serializer uint32
	Error      []byte

	Data []byte
}
