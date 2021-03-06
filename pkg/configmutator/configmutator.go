package configmutator

import (
	"encoding/json"
)

type EventConfig struct {
	Name   string         `json:"name"`
	ID     int            `json:"id"`
	Fields map[string]int `json:"fields"`
}

type MetricConfigJSON struct {
	ID         int           `json:"id"`
	Name       string        `json:"name"`
	EventIDs   []int         `json:"event_ids"`
	KeyField   string        `json:"key_field"`
	CountField string        `json:"count_field"`
	Type       string        `json:"type"`
	Filter     []interface{} `json:"filter"`
}

type MetricConfig struct {
	ID           int           `json:"id"`
	Name         string        `json:"name"`
	EventIDs     []int         `json:"event_ids"`
	KeyField     string        `json:"key_field"`
	CountField   string        `json:"count_field"`
	Type         string        `json:"type"`
	Filter       []interface{} `json:"filter"`
	FilterString string        `json:"filter_string"`
}

type Config struct {
	Namespace string
	Metrics   map[int]MetricConfig
	Events    map[int]EventConfig
	ExtraInfo map[string]interface{}
}

type ConfigJSON struct {
	Namespace string                 `json:"namespace"`
	Metrics   []MetricConfigJSON     `json:"metrics"`
	Events    []EventConfig          `json:"events"`
	ExtraInfo map[string]interface{} `json:"extra_info"`
}

type ConfigMutator struct {
	C Config
	// Data structures for easy access to common data
	AllFields       map[string]int
	KeyFields       map[string]bool
	CountFields     map[string]bool
	fieldToEventIDs map[string]map[int]bool
	nextEventID     int
	nextMetricID    int
}

func NewConfigMutator(cfg string) ConfigMutator {
	cj := ConfigJSON{}
	err := json.Unmarshal([]byte(cfg), &cj)
	if err != nil {
		panic(err)
	}
	c := Config{
		Namespace: cj.Namespace,
		Metrics:   map[int]MetricConfig{},
		Events:    map[int]EventConfig{},
		ExtraInfo: cj.ExtraInfo,
	}
	cm := ConfigMutator{}

	// Events
	for _, e := range cj.Events {
		if _, exists := c.Events[e.ID]; exists {
			panic("Duplicate event IDs")
		}
		c.Events[e.ID] = e
		if e.ID > cm.nextEventID {
			cm.nextEventID = e.ID
		}
	}
	cm.nextEventID++

	// Metrics
	for _, m := range cj.Metrics {
		if _, exists := c.Metrics[m.ID]; exists {
			panic("Duplicate metric IDs")
		}
		filterString, _ := json.Marshal(m.Filter)
		c.Metrics[m.ID] = MetricConfig{
			ID:           m.ID,
			Name:         m.Name,
			EventIDs:     m.EventIDs,
			KeyField:     m.KeyField,
			CountField:   m.CountField,
			Type:         m.Type,
			Filter:       m.Filter,
			FilterString: string(filterString),
		}
		if m.ID > cm.nextMetricID {
			cm.nextMetricID = m.ID
		}
	}
	cm.nextMetricID++
	cm.C = c
	cm.Update()
	return cm
}

func (cm *ConfigMutator) Update() {
	cm.AllFields = map[string]int{}
	cm.KeyFields = map[string]bool{}
	cm.CountFields = map[string]bool{}
	cm.fieldToEventIDs = map[string]map[int]bool{}
	// Get all fields from events, and map to ids
	for _, e := range cm.C.Events {
		for fieldName, fieldType := range e.Fields {
			// All fields
			if existingType, exists := cm.AllFields[fieldName]; exists && (existingType != fieldType) {
				panic("Conflicting field types found")
			}
			cm.AllFields[fieldName] = fieldType
			// Map to ids
			if _, exists := cm.fieldToEventIDs[fieldName]; !exists {
				cm.fieldToEventIDs[fieldName] = map[int]bool{}
			}
			cm.fieldToEventIDs[fieldName][e.ID] = true
		}
	}
	// Key fields can only be ints (for now)
	for fieldName, fieldType := range cm.AllFields {
		if fieldType == 1 {
			cm.CountFields[fieldName] = true
		}
	}
	// Count fields can only be ints (for now)
	for fieldName, fieldType := range cm.AllFields {
		if fieldType == 1 {
			cm.KeyFields[fieldName] = true
		}
	}
}
