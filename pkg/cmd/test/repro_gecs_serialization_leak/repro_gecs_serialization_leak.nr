package main

import "io"
import "serialize"
import "json"

type MyComponent = struct {
    value: i32
}

type Container = struct {
    comp: MyComponent
}

pub fn nr_deserialize_json_MyComponent(data: str) @MyComponent {
    var v_res = json.Parse(data)
    if (v_res.IsErr()) { return MyComponent { value: 0 } }
    var v = v_res.Unwrap()
    return nr_deserialize_json_MyComponent_from_val(#v)
}

pub fn nr_deserialize_json_MyComponent_from_val(v: #json.JsonValue) @MyComponent {
    var obj = v.GetObject()
    var f_value = 0
    var i = 0
    while (i < obj.Len()) {
        var prop = obj.Get(i)
        if (prop.name == "value") {
            f_value = i32(prop.value.GetNumber())
        }
        i = i + 1
    }
    return MyComponent {
        value: f_value,
    }
}

fn store_component(comp: MyComponent) Container {
    var res = Container { comp: comp }
    return res
}

fn main() i32 {
    var data = "{\"value\": 100}"
    
    // val is a heap-allocated pointer of type @MyComponent
    var val = nr_deserialize_json_MyComponent(data)
    
    // We move @val into store_component which takes MyComponent by value.
    // The pointer block val should be freed, since store_component only copies the struct.
    var container = store_component(@val)
    io.PrintLn("Value: ${container.comp.value}")
    
    return 0
}
