package aggregator

type StorageResult struct {
	Value   int
	ErrCode int
}

type AbstractStorage interface {
	Get(string) StorageResult
	Put(string, int)
	Lock()
	Unlock()
}
