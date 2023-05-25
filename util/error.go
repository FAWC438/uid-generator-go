package util

type UidGeneratorError struct {
	Msg string
}

type BufferFullError struct {
	Msg string
}

func (a *UidGeneratorError) Error() string {
	return a.Msg
}

func (a *BufferFullError) Error() string {
	return a.Msg
}
