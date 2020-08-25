package aggregator

type AbstractFilter interface {
    IsValid(Event) bool
}
