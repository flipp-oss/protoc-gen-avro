# protoc-gen-avro

Generate Avro schemas from Protobuf files.

## Usage

### Install using Golang 1.21+

Simply run the following command:

```bash
go install github.com/flipp-oss/protoc-gen-avro@latest
```

### Install Manually

Download this project from the Releases page. Put the generated binary in your
path:

```bash
mv protoc-gen-avro /usr/local/bin
```

Generate Avro schemas from your Protobuf files:

```bash
protoc --avro_out=. *.proto
```

## Options

- `emit_only` - A semicolon-separated list of records to emit. If not specified,
  all records will be emitted.

```bash
protoc --avro_out=. --avro_opt=emit_only=Foo;Bar *.proto
```

This will generate only `Foo.avsc` and `Bar.avsc` files.

- `namespace_map` - A comma-separated list of namespaces to map. If not
  specified, all namespaces will be mapped.

```bash
protoc --avro_out=. --avro_opt=namespace_map=foo:bar,baz:spam *.proto
```

...will change the output namespace for `foo` to `bar` and `baz` to `spam`.

- `collapse_fields` - A semicolon-separated list of records to collapse.
  Collapsed records should have a single field in them, and they will be
  replaced in the output by that field. This can be useful to overcome some
  limitations of Protobuf - e.g. Protobuf doesn't have the ability to have an
  array of maps, while Avro does.

```bash
protoc --avro_out=. --avro_opt=collapse_fields=StringList;SomeOtherRecord *.proto
```

If you have the input proto looking like:

```protobuf
message StringList {
  repeated string strings = 1;
}
message MyRecord {
    map<string, StringList> my_field = 1;
}
```

...the output will look like:

```json
{
  "type": "record",
  "name": "MyRecord",
  "fields": [
    {
      "name": "my_field",
      "type": {
        "type": "map",
        "values": {
          "type": {
            "type": "array",
            "items": "string"
          }
        }
      }
    }
  ]
}
```

- `remove_enum_prefixes` - if set to true, will remove the prefixes from enum
  values. E.g. if you have an enum like:

```protobuf
enum Category {
  CATEGORY_GOOD = 0;
  CATEGORY_BAD = 1;
}
```

...with this option on, it will be translated into:

```json
{
  "type": "enum",
  "name": "CATEGORY",
  "symbols": ["GOOD", "BAD"]
}
```

- `preserve_non_string_maps` - if set to true, will replace maps with non-string
  keys with records. E.g. if you have a map like:

```protobuf
message MyRecord {
  map<int32, string> my_field = 1;
}
```

...with this option off, it will be translated into:

```json
{
  "type": "record",
  "name": "MyRecord",
  "fields": [
    {
      "name": "my_field",
      "type": {
        "type": "map",
        "values": "string"
      }
    }
  ]
}
```

...but with it on, it will be translated into:

```json
{
  "type": "record",
  "name": "MyRecord",
  "fields": [
    {
      "name": "my_field",
      "type": {
        "type": "record",
        "fields": [
          {
            "name": "key",
            "type": "int"
          },
          {
            "name": "value",
            "type": "string"
          }
        ]
      }
    }
  ]
}
```

---

To Do List:

- Add tests
- Homebrew?

---

This project supported by [Flipp](https://corp.flipp.com/).
