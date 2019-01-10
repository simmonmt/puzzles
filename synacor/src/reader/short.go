package reader

type Short interface {
	Read() (uint16, error)
	Off() uint16
}
