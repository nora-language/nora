package main

import "crypto"
import "io"
import "string"

fn assert(cond: bool, msg: str) {
    if (!cond) {
        panic(msg)
    }
}

fn assert_string_eq(s1: str, s2: str, msg: str) {
    assert(string.Equals(s1, s2), msg + " -> Expected: " + s2 + ", Got: " + s1)
}

fn test_hex() {
    io.PrintLn("Running Hex tests...")

    // 1. Encoding
    var hello_bytes = crypto.StringToBytes("hello")
    var hex_str = crypto.EncodeHex(hello_bytes)
    assert_string_eq(hex_str, "68656c6c6f", "Hex encoding of 'hello' failed")

    // 2. Decoding
    match crypto.DecodeHex("68656c6c6f") {
        Ok(decoded) => {
            var decoded_str = crypto.BytesToString(decoded)
            assert_string_eq(decoded_str, "hello", "Hex decoding of 'hello' failed")
        }
        Err(err) => {
            panic("Hex decoding failed with error: " + err)
        }
    }

    // 3. Decoding Errors
    match crypto.DecodeHex("123") {
        Ok(val) => {
            panic("Hex decoding should fail for uneven length string")
        }
        Err(err) => {
            io.PrintLn("  (Pass) Correctly caught uneven length: " + err)
        }
    }

    match crypto.DecodeHex("123g") {
        Ok(val) => {
            panic("Hex decoding should fail for invalid characters")
        }
        Err(err) => {
            io.PrintLn("  (Pass) Correctly caught invalid character: " + err)
        }
    }

    io.PrintLn("Hex tests passed!")
}

fn test_base64() {
    io.PrintLn("Running Base64 tests...")

    // 1. Encoding
    var empty_bytes = crypto.StringToBytes("")
    var empty_b64 = crypto.EncodeBase64(empty_bytes)
    assert_string_eq(empty_b64, "", "Base64 empty failed")

    var foo_bytes = crypto.StringToBytes("foo")
    var foo_b64 = crypto.EncodeBase64(foo_bytes)
    assert_string_eq(foo_b64, "Zm9v", "Base64 'foo' failed")

    var foobar_bytes = crypto.StringToBytes("foobar")
    var foobar_b64 = crypto.EncodeBase64(foobar_bytes)
    assert_string_eq(foobar_b64, "Zm9vYmFy", "Base64 'foobar' failed")

    // 2. Decoding
    match crypto.DecodeBase64("Zm9v") {
        Ok(dec) => {
            var s = crypto.BytesToString(dec)
            assert_string_eq(s, "foo", "Base64 decoding of 'foo' failed")
        }
        Err(err) => {
            panic("Base64 decoding failed for 'Zm9v'")
        }
    }

    match crypto.DecodeBase64("Zm9vYmFy") {
        Ok(dec) => {
            var s = crypto.BytesToString(dec)
            assert_string_eq(s, "foobar", "Base64 decoding of 'foobar' failed")
        }
        Err(err) => {
            panic("Base64 decoding failed for 'Zm9vYmFy'")
        }
    }

    // 3. Decoding Errors
    match crypto.DecodeBase64("Zm9") {
        Ok(val) => {
            panic("Base64 decoding should fail for invalid length")
        }
        Err(err) => {
            io.PrintLn("  (Pass) Correctly caught invalid length: " + err)
        }
    }

    match crypto.DecodeBase64("Zm9v*mFy") {
        Ok(val) => {
            panic("Base64 decoding should fail for invalid characters")
        }
        Err(err) => {
            io.PrintLn("  (Pass) Correctly caught invalid character: " + err)
        }
    }

    io.PrintLn("Base64 tests passed!")
}

fn test_sha256() {
    io.PrintLn("Running SHA-256 tests...")

    // 1. Empty string hash
    var hash1 = crypto.Sha256String("")
    assert_string_eq(hash1, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "SHA-256 of empty string failed")

    // 2. Short string hash ("abc")
    var hash2 = crypto.Sha256String("abc")
    assert_string_eq(hash2, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad", "SHA-256 of 'abc' failed")

    // 3. Long string without period
    var hash3 = crypto.Sha256String("The quick brown fox jumps over the lazy dog")
    assert_string_eq(hash3, "d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592", "SHA-256 of long string without period failed")

    // 4. Long string with period
    var hash4 = crypto.Sha256String("The quick brown fox jumps over the lazy dog.")
    assert_string_eq(hash4, "ef537f25c895bfa782526529a9b63d97aa631564d5d789c2b765448c8635fb6c", "SHA-256 of long string with period failed")

    io.PrintLn("SHA-256 tests passed!")
}

fn test_secure_random() {
    io.PrintLn("Running Secure Random tests...")

    // 1. RandBytes
    var bytes1 = crypto.RandBytes(16)
    assert(len(bytes1) == 16, "RandBytes(16) should return 16 bytes")

    var bytes2 = crypto.RandBytes(16)
    assert(len(bytes2) == 16, "RandBytes(16) should return 16 bytes")

    // Verify distinctness (extremely high probability they are not identical)
    var is_identical = true
    var i = 0
    while (i < 16) {
        if (bytes1[i] != bytes2[i]) {
            is_identical = false
        }
        i = i + 1
    }
    assert(!is_identical, "RandBytes(16) consecutive calls should generate distinct bytes")

    // 2. RandInt
    var r1 = crypto.RandInt()
    var r2 = crypto.RandInt()
    assert(r1 != r2, "RandInt() consecutive calls should generate distinct values")

    // 3. RandIntRange
    var min = 50
    var max = 150
    i = 0
    while (i < 100) {
        var val = crypto.RandIntRange(min, max)
        assert(val >= min && val < max, "RandIntRange out of bounds")
        i = i + 1
    }

    io.PrintLn("Secure Random tests passed!")
}

fn main() i32 {
    io.PrintLn("=== Starting Nora Premium Cryptographic Standard Library Tests ===")

    test_hex()
    test_base64()
    test_sha256()
    test_secure_random()

    io.PrintLn("=== All Cryptographic Tests Passed Successfully! ===")
    return 0
}
