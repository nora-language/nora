package main

import "io"
import "serialize"
import "string"

[serialize]
type Address = struct {
    street: str,
    zip: i32
}

[serialize]
type Item = struct {
    name: str,
    price: f64
}

[serialize]
type User = struct {
    id: i32,
    [rename("full_name")]
    name: str,
    address: Address,
    scores: i32[],
    is_active: bool,
    ratio: f64,
    factor: f32,
    big_id: i64,
    tags: str[],
    items: Item[],
    
    // Borrowed fields must be ignored by the compiler-generated serializer
    borrowed_address: #Address,
    mut_borrowed_val: &i32
}

fn main() i32 {
    io.PrintLn("--- Testing Serialization ---")

    var addr_for_borrow = Address { street: "Second St", zip: 54321 }
    var val_for_mut_borrow = 456

    // --- Scenario 1: Standard Serialization & Deserialization ---
    io.PrintLn("Scenario 1: Testing standard serialization with all supported fields")
    
    var scores = alloc i32[3]
    scores[0] = 95
    scores[1] = 88
    scores[2] = 100

    var tags = alloc str[2]
    tags[0] = "admin"
    tags[1] = "staff"

    var items = alloc Item[2]
    items[0] = Item { name: "Laptop", price: 1200.50 }
    items[1] = Item { name: "Mouse", price: 25.75 }

    var u = User {
        id: 42,
        name: "Alice",
        address: Address {
            street: "Main St",
            zip: 12345
        },
        scores: @scores,
        is_active: true,
        ratio: 3.14159,
        factor: f32(1.5),
        big_id: i64(987654321),
        tags: @tags,
        items: @items,
        borrowed_address: #addr_for_borrow,
        mut_borrowed_val: &val_for_mut_borrow
    }

    var json_str = serialize.ToJSON[User](#u)
    io.PrintLn("Stringified JSON:")
    io.PrintLn(json_str)

    // Verify key renaming
    if (!string.Contains(json_str, "\"full_name\":\"Alice\"")) {
        panic("expected renamed field key 'full_name' in JSON")
    }
    if (string.Contains(json_str, "\"name\":\"Alice\"")) {
        panic("should not contain original field key 'name' when renamed")
    }

    // Verify borrowed fields are ignored (not serialized)
    if (string.Contains(json_str, "borrowed_address")) {
        panic("borrowed fields should be ignored and not present in JSON")
    }
    if (string.Contains(json_str, "mut_borrowed_val")) {
        panic("mutable borrowed fields should be ignored and not present in JSON")
    }

    var u2 = serialize.FromJSON[User](json_str)
    io.PrintLn("Deserialized User u2 values verified.")

    if (u2.id != 42) { panic("expected ID to be 42") }
    if (u2.name != "Alice") { panic("expected Name to be Alice") }
    if (u2.address.street != "Main St") { panic("expected Street to be Main St") }
    if (u2.address.zip != 12345) { panic("expected ZIP to be 12345") }
    if (len(u2.scores) != 3) { panic("expected 3 scores") }
    if (u2.scores[0] != 95) { panic("expected first score to be 95") }
    if (u2.scores[1] != 88) { panic("expected second score to be 88") }
    if (u2.scores[2] != 100) { panic("expected third score to be 100") }
    if (u2.is_active != true) { panic("expected is_active to be true") }
    if (u2.ratio != 3.14159) { panic("expected ratio to be 3.14159") }
    if (u2.factor != f32(1.5)) { panic("expected factor to be 1.5") }
    if (u2.big_id != i64(987654321)) { panic("expected big_id to be 987654321") }
    if (len(u2.tags) != 2) { panic("expected 2 tags") }
    if (u2.tags[0] != "admin") { panic("expected first tag to be admin") }
    if (u2.tags[1] != "staff") { panic("expected second tag to be staff") }
    if (len(u2.items) != 2) { panic("expected 2 items") }
    if (u2.items[0].name != "Laptop") { panic("expected item 0 name to be Laptop") }
    if (u2.items[0].price != 1200.50) { panic("expected item 0 price to be 1200.50") }
    if (u2.items[1].name != "Mouse") { panic("expected item 1 name to be Mouse") }
    if (u2.items[1].price != 25.75) { panic("expected item 1 price to be 25.75") }

    // --- Scenario 2: Special Characters in Strings ---
    io.PrintLn("Scenario 2: Testing string escaping with special characters")
    
    var empty_scores = alloc i32[0]
    var empty_tags = alloc str[0]
    var empty_items = alloc Item[0]

    var u_special = User {
        id: 101,
        name: "Bob\n\"The Great\"\t\\\r",
        address: Address { street: "Escaped Ln", zip: 999 },
        scores: @empty_scores,
        is_active: false,
        ratio: 0.0,
        factor: f32(0.0),
        big_id: i64(0),
        tags: @empty_tags,
        items: @empty_items,
        borrowed_address: #addr_for_borrow,
        mut_borrowed_val: &val_for_mut_borrow
    }

    var special_json = serialize.ToJSON[User](#u_special)
    var u_special_deserialized = serialize.FromJSON[User](special_json)

    if (u_special_deserialized.name != "Bob\n\"The Great\"\t\\\r") {
        panic("expected special character string to be correctly escaped and round-tripped")
    }

    // --- Scenario 3: Partial JSON Deserialization ---
    io.PrintLn("Scenario 3: Testing fallback to default values on partial JSON payloads")
    
    var partial_json = "{\"id\":777,\"full_name\":\"Charlie\"}"
    var u_partial = serialize.FromJSON[User](partial_json)

    if (u_partial.id != 777) { panic("expected partial ID to be 777") }
    if (u_partial.name != "Charlie") { panic("expected partial Name to be Charlie") }
    // All other fields should resolve to their respective default values
    if (u_partial.address.street != "") { panic("expected default empty street") }
    if (u_partial.address.zip != 0) { panic("expected default ZIP of 0") }
    if (len(u_partial.scores) != 0) { panic("expected default empty scores array") }
    if (u_partial.is_active != false) { panic("expected default is_active to be false") }
    if (u_partial.ratio != 0.0) { panic("expected default ratio of 0.0") }
    if (u_partial.factor != f32(0.0)) { panic("expected default factor of 0.0") }
    if (u_partial.big_id != i64(0)) { panic("expected default big_id of 0") }
    if (len(u_partial.tags) != 0) { panic("expected default empty tags array") }
    if (len(u_partial.items) != 0) { panic("expected default empty items array") }

    // --- Scenario 4: Corrupted/Invalid JSON ---
    io.PrintLn("Scenario 4: Testing fallback to defaults on corrupted/invalid JSON inputs")
    
    var corrupted_json = "{invalid_json_payload:"
    var u_corrupted = serialize.FromJSON[User](corrupted_json)

    if (u_corrupted.id != 0) { panic("expected corrupted fallback ID to be 0") }
    if (u_corrupted.name != "") { panic("expected corrupted fallback Name to be empty") }
    if (u_corrupted.address.street != "") { panic("expected corrupted fallback street to be empty") }
    if (u_corrupted.address.zip != 0) { panic("expected corrupted fallback ZIP to be 0") }
    if (len(u_corrupted.scores) != 0) { panic("expected corrupted fallback scores array to be empty") }

    io.PrintLn("--- Serialization Test Passed Successfully! ---")
    return 0
}
