package avro

import (
  "fmt"
  "github.com/iancoleman/orderedmap"
  "reflect"
  "slices"
)

type Union struct {
  Types []Type
}

func (u Union) ToJSON(types *TypeRepo) (any, error) {
  var jsonSlice []any
  for _, unionType := range u.Types {
    typeJson, err := unionType.ToJSON(types)
    if err != nil {
      return nil, fmt.Errorf("error parsing union type: %w", err)
    }
    jsonSlice = append(jsonSlice, typeJson)
  }
  return flatten(jsonSlice), nil
}

func flatten(slice []any) []any {
  var flattened []any
  for _, jsonType := range slice {
    jsonMap, ok := jsonType.(*orderedmap.OrderedMap)
    if ok {
      typeArr, ok := jsonMap.Get("type")
      if ok && reflect.TypeOf(typeArr).Kind() == reflect.Slice {
        for _, nestedType := range typeArr.([]any) {
          flattened = appendUnique(flattened, nestedType)
        }
      } else {
        flattened = append(flattened, jsonType)
      }
    } else {
      flattened = appendUnique(flattened, jsonType)
    }
  }
  return flattened
}

// Avro unions can't contain the same type twice - e.g. an optional oneof would otherwise
// give ["null", "null", ...].
func appendUnique(slice []any, jsonType any) []any {
  name, ok := jsonType.(string)
  if ok && slices.Contains(slice, any(name)) {
    return slice
  }
  return append(slice, jsonType)
}
