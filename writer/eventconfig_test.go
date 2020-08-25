package aggregator_test

import (
	"testing"

	. "github.com/kal-g/aggregator-go/writer"
)

func TestEventIdValidation(t *testing.T) {
	ec := EventConfig{
		Name:   "testEvent",
		Id:     1,
		Fields: map[string]FieldType{},
	}

	// Create events
	re1 := map[string]interface{}{"id": 1}
	re2 := map[string]interface{}{"id": 2}

	// Validate events
	v1 := ec.Validate(re1)
	v2 := ec.Validate(re2)

	var nil_event *Event
	valid_event := new(Event)
	valid_event.Id = 1

	AssertEqual(t, v1.Id, valid_event.Id)
	for k, v := range v1.Data {
		AssertEqual(t, v, valid_event.Data[k])
	}
	AssertEqual(t, v2, nil_event)
}

/*
TEST_CASE("Validate number of fields") {
  // Create config
  std::unordered_map<std::string, FieldType> fields(
      {{"testField1", StringField}});
  EventConfig testConfig("testEvent", 1, std::move(fields));

  // Create event to validate
  RawEvent validEvent({{"id", 1}, {"testField1", "testValue"}});
  RawEvent invalidEvent(
      {{"id", 1}, {"testField1", "testValue"}, {"testField2", "testValue"}});

  REQUIRE(testConfig.validate(std::move(validEvent)));
  REQUIRE(!testConfig.validate(std::move(invalidEvent)));
}
*/

func TestValidateNumberOfFields(t *testing.T) {
	// Create config
	ec := EventConfig{
		Name: "testEvent",
		Id:   1,
		Fields: map[string]FieldType{
			"testField1": StringField,
		},
	}

	// Create event to validate
	valid_event := map[string]interface{}{
		"id":         1,
		"testField1": "testValue",
	}
	invalid_event := map[string]interface{}{
		"id":         1,
		"testField1": "testValue",
		"testField2": "testValue",
	}

	// Check
	var nil_event *Event
	v1 := ec.Validate(valid_event)
	v2 := ec.Validate(invalid_event)

	valid_event_e := new(Event)
	valid_event_e.Id = 1
	valid_event_e.Data = map[string]interface{}{}

	for k, v := range valid_event {
		valid_event_e.Data[k] = v
	}
	delete(valid_event_e.Data, "id")

	AssertEqual(t, v1, valid_event_e)
	AssertEqual(t, v2, nil_event)
}

func TestValidateFieldTypes(t *testing.T) {
	ec := EventConfig{
		Name: "testEvent",
		Id:   1,
		Fields: map[string]FieldType{
			"testField1": StringField,
			"testField2": StringField,
			"testField3": IntField,
		},
	}

	valid_event := map[string]interface{}{
		"id":         1,
		"testField1": "testValue",
		"testField2": "testValue",
		"testField3": 0,
	}
	invalid_event_1 := map[string]interface{}{
		"id":         1,
		"testField1": "testValue",
		"testField2": 0,
		"testField3": 0,
	}
	invalid_event_2 := map[string]interface{}{
		"id":         1,
		"testField1": "testValue",
		"testField2": "testValue",
		"testField3": "testValue",
	}

	// Validate events
	v1 := ec.Validate(valid_event)
	v2 := ec.Validate(invalid_event_1)
	v3 := ec.Validate(invalid_event_2)

	var nil_event *Event
	valid_event_e := new(Event)
	valid_event_e.Id = 1
	valid_event_e.Data = map[string]interface{}{}

	for k, v := range valid_event {
		valid_event_e.Data[k] = v
	}
	delete(valid_event_e.Data, "id")

	AssertEqual(t, v1.Id, valid_event_e.Id)
	for k, v := range v1.Data {
		AssertEqual(t, v, valid_event_e.Data[k])

	}
	AssertEqual(t, v2, nil_event)
	AssertEqual(t, v3, nil_event)
}
