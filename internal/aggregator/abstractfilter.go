package aggregator

type abstractFilter interface {
	IsValid(event) bool
}
