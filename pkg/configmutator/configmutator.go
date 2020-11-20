package configmutator

import (
	"encoding/json"
)

type EventConfig struct {
	Name   string         `json:"name"`
	ID     int            `json:"id"`
	Fields map[string]int `json:"fields"`
}

type MetricConfig struct {
	ID         int           `json:"id"`
	Name       string        `json:"name"`
	EventIDs   []int         `json:"event_ids"`
	KeyField   string        `json:"key_field"`
	CountField string        `json:"count_field"`
	Type       string        `json:"type"`
	Filter     []interface{} `json:"filter"`
}

type Config struct {
	Namespace string
	Metrics   map[int]MetricConfig
	Events    map[int]EventConfig
	ExtraInfo map[string]interface{}
}

type ConfigJSON struct {
	Namespace string                 `json:"namespace"`
	Metrics   []MetricConfig         `json:"metrics"`
	Events    []EventConfig          `json:"events"`
	ExtraInfo map[string]interface{} `json:"extra_info"`
}

type ConfigMutator struct {
	c Config
	// Data structures for easy access to common data
	allFields       map[string]int
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
		c.Metrics[m.ID] = m
		if m.ID > cm.nextMetricID {
			cm.nextMetricID = m.ID
		}
	}
	cm.nextMetricID++
	cm.c = c
	cm.Update()
	return cm
}

func (cm *ConfigMutator) Update() {
	cm.allFields = map[string]int{}
	cm.KeyFields = map[string]bool{}
	cm.CountFields = map[string]bool{}
	cm.fieldToEventIDs = map[string]map[int]bool{}
	// Get all fields from events, and map to ids
	for _, e := range cm.c.Events {
		for fieldName, fieldType := range e.Fields {
			// All fields
			if existingType, exists := cm.allFields[fieldName]; exists && (existingType != fieldType) {
				panic("Conflicting field types found")
			}
			cm.allFields[fieldName] = fieldType
			// Map to ids
			if _, exists := cm.fieldToEventIDs[fieldName]; !exists {
				cm.fieldToEventIDs[fieldName] = map[int]bool{}
			}
			cm.fieldToEventIDs[fieldName][e.ID] = true
		}
	}
	// Key fields can only be ints (for now)
	for fieldName, fieldType := range cm.allFields {
		if fieldType == 1 {
			cm.CountFields[fieldName] = true
		}
	}
	// Count fields can only be ints (for now)
	for fieldName, fieldType := range cm.allFields {
		if fieldType == 1 {
			cm.KeyFields[fieldName] = true
		}
	}
}
