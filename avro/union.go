package avro

import (
  "fmt"
  "github.com/iancoleman/orderedmap"
  "reflect"
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
    LogMsg("%v", reflect.TypeOf(jsonType))
    if ok {
      typeArr, ok := jsonMap.Get("type")
      if ok && reflect.TypeOf(typeArr).Kind() == reflect.Slice {
        flattened = append(flattened, typeArr.([]any)...)
      } else {
        flattened = append(flattened, jsonType)
      }
    } else {
      flattened = append(flattened, jsonType)
    }
  }
  return flattened
}
