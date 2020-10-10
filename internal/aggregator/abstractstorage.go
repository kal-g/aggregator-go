package aggregator

// StorageResult is the wrapper for getting values from storagge
type StorageResult struct {
	Value int
	Err   error
}

// AbstractStorage is the storage interface
// Should be threadsafe via lock / unlock
type AbstractStorage interface {
	Get(string) StorageResult
	IncrBy(string, int)
	Lock(string)
	Unlock(string)
}
