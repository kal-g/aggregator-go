package aggregator

import (
	"testing"

	ct "github.com/kal-g/aggregator-go/internal/common_test"
)

func TestEventIdValidation(t *testing.T) {
	ec := eventConfig{
		Name:   "testEvent",
		ID:     1,
		Fields: map[string]fieldType{},
	}

	// Create events
	re1 := map[string]interface{}{"id": 1}
	re2 := map[string]interface{}{"id": 2}

	// Validate events
	v1 := ec.validate(re1)
	v2 := ec.validate(re2)

	var nilEvent *event
	validEvent := new(event)
	validEvent.ID = 1

	ct.AssertEqual(t, v1.ID, validEvent.ID)
	for k, v := range v1.Data {
		ct.AssertEqual(t, v, validEvent.Data[k])
	}
	ct.AssertEqual(t, v2, nilEvent)
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
	ec := eventConfig{
		Name: "testEvent",
		ID:   1,
		Fields: map[string]fieldType{
			"testField1": stringField,
		},
	}

	// Create event to validate
	validEvent := map[string]interface{}{
		"id":         1,
		"testField1": "testValue",
	}
	invalidEvent := map[string]interface{}{
		"id":         1,
		"testField1": "testValue",
		"testField2": "testValue",
	}

	// Check
	var nilEvent *event
	v1 := ec.validate(validEvent)
	v2 := ec.validate(invalidEvent)

	validEventE := new(event)
	validEventE.ID = 1
	validEventE.Data = map[string]interface{}{}

	for k, v := range validEvent {
		validEventE.Data[k] = v
	}
	delete(validEventE.Data, "id")

	ct.AssertEqual(t, v1, validEventE)
	ct.AssertEqual(t, v2, nilEvent)
}

func TestValidateFieldTypes(t *testing.T) {
	ec := eventConfig{
		Name: "testEvent",
		ID:   1,
		Fields: map[string]fieldType{
			"testField1": stringField,
			"testField2": stringField,
			"testField3": intField,
		},
	}

	validEvent := map[string]interface{}{
		"id":         1,
		"testField1": "testValue",
		"testField2": "testValue",
		"testField3": 0,
	}
	invalidEvent1 := map[string]interface{}{
		"id":         1,
		"testField1": "testValue",
		"testField2": 0,
		"testField3": 0,
	}
	invalidEvent2 := map[string]interface{}{
		"id":         1,
		"testField1": "testValue",
		"testField2": "testValue",
		"testField3": "testValue",
	}

	// Validate events
	v1 := ec.validate(validEvent)
	v2 := ec.validate(invalidEvent1)
	v3 := ec.validate(invalidEvent2)

	var nilEvent *event
	validEventE := new(event)
	validEventE.ID = 1
	validEventE.Data = map[string]interface{}{}

	for k, v := range validEvent {
		validEventE.Data[k] = v
	}
	delete(validEventE.Data, "id")

	ct.AssertEqual(t, v1.ID, validEventE.ID)
	for k, v := range v1.Data {
		ct.AssertEqual(t, v, validEventE.Data[k])

	}
	ct.AssertEqual(t, v2, nilEvent)
	ct.AssertEqual(t, v3, nilEvent)
}
