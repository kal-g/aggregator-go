package configvalidator

import (
	"encoding/json"
)

func ValidateFilter(filterList []interface{}, fieldsByType map[string]int) bool {
	if len(filterList) == 0 {
		return false
	}

	if filterList[0] == "null" {
		return len(filterList) == 1
	} else if filterList[0] == "gt" {
		if len(filterList) != 3 {
			return false
		}
		fieldNameString, isString := filterList[1].(string)
		if !isString {
			return false
		}
		fieldType, fieldExists := fieldsByType[fieldNameString]
		if !fieldExists {
			return false
		}
		if fieldType != 1 {
			return false
		}
	} else if filterList[0] == "all" {
		if len(filterList) < 2 {
			return false
		}
		for i := 1; i < len(filterList); i++ {
			nFilterList, isList := filterList[i].([]interface{})
			if !isList {
				return false
			}
			if !ValidateFilter(nFilterList, fieldsByType) {
				return false
			}
		}
	} else {
		return false
	}

	return true
}

func ValidateFilterString(fString string, fieldsByType map[string]int) bool {
	filterList := []interface{}{}
	err := json.Unmarshal([]byte(fString), &filterList)
	if err != nil {
		return false
	}

	return ValidateFilter(filterList, fieldsByType)
}
