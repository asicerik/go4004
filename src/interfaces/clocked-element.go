package interfaces

type ClockedElement interface {
	Reset()
	Clock()
}
