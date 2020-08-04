package datastore

type datastoreError string

func (dse datastoreError) Error() string {
	return string(dse)
}

// Datastore errors
const (
	ErrIncompatibleTransaction datastoreError = "Existing transaction's writable state is not compatible"
)
