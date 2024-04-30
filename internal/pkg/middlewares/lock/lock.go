package lock

type Lock interface {
	Lock()
	Unlock()
}
