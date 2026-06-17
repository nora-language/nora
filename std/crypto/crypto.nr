package crypto

extern fn strlen(s: #str) i32

// --- Utility Helpers ---
pub fn StringToBytes(s: #str) byte[0] {
    var len_s = strlen(s)
    var b = alloc byte[len_s]
    var i = 0
    while i < len_s {
        b[i] = s[i]
        i = i + 1
    }
    return b
}

pub fn BytesToString(b: #byte[0]) str {
    var len_b = len(b)
    var s = alloc byte[len_b + 1]
    var i = 0
    while i < len_b {
        s[i] = b[i]
        i = i + 1
    }
    s[len_b] = 0
    return s
}

// --- Hex Encoding / Decoding ---
pub fn EncodeHex(data: #byte[0]) str {
    var hex_chars = "0123456789abcdef"
    var len_data = len(data)
    var b = alloc byte[len_data * 2 + 1]
    var i = 0
    while i < len_data {
        var val: i32 = data[i]
        b[i * 2] = hex_chars[(val >> 4) & 15]
        b[i * 2 + 1] = hex_chars[val & 15]
        i = i + 1
    }
    b[len_data * 2] = 0
    return b
}

fn get_hex_val(c: byte) i32 {
    var val: i32 = c
    if (val >= 48 && val <= 57) {
        return val - 48
    }
    if (val >= 97 && val <= 102) {
        return val - 97 + 10
    }
    if (val >= 65 && val <= 70) {
        return val - 65 + 10
    }
    return -1
}

pub fn DecodeHex(s: #str) Result[byte[0], str] {
    var len_s = strlen(s)
    if (len_s % 2 != 0) {
        return Err[byte[0], str]("hex string must have an even length")
    }
    var num_bytes = len_s / 2
    var b = alloc byte[num_bytes]
    var i = 0
    while i < num_bytes {
        var h1 = get_hex_val(s[i * 2])
        var h2 = get_hex_val(s[i * 2 + 1])
        if (h1 < 0 || h2 < 0) {
            return Err[byte[0], str]("invalid hex character")
        }
        b[i] = (h1 << 4) | h2
        i = i + 1
    }
    return Ok[byte[0], str](b)
}

// --- Base64 Encoding / Decoding ---
pub fn EncodeBase64(data: #byte[0]) str {
    var b64_chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
    var len_data = len(data)

    var num_groups = len_data / 3
    var rem = len_data % 3
    var output_len = num_groups * 4
    if (rem != 0) {
        output_len = output_len + 4
    }

    var b = alloc byte[output_len + 1]
    var src_idx = 0
    var dst_idx = 0

    while src_idx < len_data - rem {
        var b0: i32 = data[src_idx]
        var b1: i32 = data[src_idx + 1]
        var b2: i32 = data[src_idx + 2]
        var val = ((b0 & 255) << 16) | ((b1 & 255) << 8) | (b2 & 255)
        b[dst_idx] = b64_chars[(val >> 18) & 63]
        b[dst_idx + 1] = b64_chars[(val >> 12) & 63]
        b[dst_idx + 2] = b64_chars[(val >> 6) & 63]
        b[dst_idx + 3] = b64_chars[val & 63]

        src_idx = src_idx + 3
        dst_idx = dst_idx + 4
    }

    if (rem == 1) {
        var b0: i32 = data[src_idx]
        var val = (b0 & 255) << 16
        b[dst_idx] = b64_chars[(val >> 18) & 63]
        b[dst_idx + 1] = b64_chars[(val >> 12) & 63]
        b[dst_idx + 2] = 61 // '='
        b[dst_idx + 3] = 61 // '='
    } else if (rem == 2) {
        var b0: i32 = data[src_idx]
        var b1: i32 = data[src_idx + 1]
        var val = ((b0 & 255) << 16) | ((b1 & 255) << 8)
        b[dst_idx] = b64_chars[(val >> 18) & 63]
        b[dst_idx + 1] = b64_chars[(val >> 12) & 63]
        b[dst_idx + 2] = b64_chars[(val >> 6) & 63]
        b[dst_idx + 3] = 61 // '='
    }

    b[output_len] = 0
    return b
}

fn get_b64_val(c: byte) i32 {
    var val: i32 = c
    if (val >= 65 && val <= 90) {
        return val - 65
    }
    if (val >= 97 && val <= 122) {
        return val - 97 + 26
    }
    if (val >= 48 && val <= 57) {
        return val - 48 + 52
    }
    if (val == 43) { // '+'
        return 62
    }
    if (val == 47) { // '/'
        return 63
    }
    return -1
}

pub fn DecodeBase64(s: #str) Result[byte[0], str] {
    var len_s = strlen(s)
    if (len_s == 0) {
        var empty = alloc byte[0]
        return Ok[byte[0], str](empty)
    }
    if (len_s % 4 != 0) {
        return Err[byte[0], str]("invalid base64 string length")
    }

    var padding = 0
    if (s[len_s - 1] == 61) { // '='
        padding = padding + 1
    }
    if (s[len_s - 2] == 61) { // '='
        padding = padding + 1
    }

    var num_bytes = (len_s / 4) * 3 - padding
    var b = alloc byte[num_bytes]

    var src_idx = 0
    var dst_idx = 0
    var limit = len_s
    if (padding > 0) {
        limit = len_s - 4
    }

    while src_idx < limit {
        var v0 = get_b64_val(s[src_idx])
        var v1 = get_b64_val(s[src_idx + 1])
        var v2 = get_b64_val(s[src_idx + 2])
        var v3 = get_b64_val(s[src_idx + 3])

        if (v0 < 0 || v1 < 0 || v2 < 0 || v3 < 0) {
            return Err[byte[0], str]("invalid base64 character")
        }

        var val = (v0 << 18) | (v1 << 12) | (v2 << 6) | v3
        b[dst_idx] = (val >> 16) & 255
        b[dst_idx + 1] = (val >> 8) & 255
        b[dst_idx + 2] = val & 255

        src_idx = src_idx + 4
        dst_idx = dst_idx + 3
    }

    if (padding == 1) {
        var v0 = get_b64_val(s[src_idx])
        var v1 = get_b64_val(s[src_idx + 1])
        var v2 = get_b64_val(s[src_idx + 2])
        if (v0 < 0 || v1 < 0 || v2 < 0) {
            return Err[byte[0], str]("invalid base64 character")
        }
        var val = (v0 << 18) | (v1 << 12) | (v2 << 6)
        b[dst_idx] = (val >> 16) & 255
        b[dst_idx + 1] = (val >> 8) & 255
    } else if (padding == 2) {
        var v0 = get_b64_val(s[src_idx])
        var v1 = get_b64_val(s[src_idx + 1])
        if (v0 < 0 || v1 < 0) {
            return Err[byte[0], str]("invalid base64 character")
        }
        var val = (v0 << 18) | (v1 << 12)
        b[dst_idx] = (val >> 16) & 255
    }

    return Ok[byte[0], str](b)
}

// --- SHA-256 Hashing ---
fn shr_logical(x: i32, n: i32) i32 {
    if (n == 0) {
        return x
    }
    var shifted = x >> n
    var mask = (1 << (32 - n)) - 1
    return shifted & mask
}

fn rotr(x: i32, n: i32) i32 {
    var left = shr_logical(x, n)
    var right = x << (32 - n)
    return left | right
}

pub fn Sha256(data: #byte[0]) byte[0] {
    var K = alloc i32[64]
    K[0] = 1116352408
    K[1] = 1899447441
    K[2] = -1245643825
    K[3] = -373957723
    K[4] = 961987163
    K[5] = 1508970993
    K[6] = -1841331548
    K[7] = -1424204075
    K[8] = -670586216
    K[9] = 310598401
    K[10] = 607225278
    K[11] = 1426881987
    K[12] = 1925078388
    K[13] = -2132889090
    K[14] = -1680079193
    K[15] = -1046744716
    K[16] = -459576895
    K[17] = -272742522
    K[18] = 264347078
    K[19] = 604807628
    K[20] = 770255983
    K[21] = 1249150122
    K[22] = 1555081692
    K[23] = 1996064986
    K[24] = -1740746414
    K[25] = -1473132947
    K[26] = -1341970488
    K[27] = -1084653625
    K[28] = -958395405
    K[29] = -710438585
    K[30] = 113926993
    K[31] = 338241895
    K[32] = 666307205
    K[33] = 773529912
    K[34] = 1294757372
    K[35] = 1396182291
    K[36] = 1695183700
    K[37] = 1986661051
    K[38] = -2117940946
    K[39] = -1838011259
    K[40] = -1564481375
    K[41] = -1474664885
    K[42] = -1035236496
    K[43] = -949202525
    K[44] = -778901479
    K[45] = -694614492
    K[46] = -200395387
    K[47] = 275423344
    K[48] = 430227734
    K[49] = 506948616
    K[50] = 659060556
    K[51] = 883997877
    K[52] = 958139571
    K[53] = 1322822218
    K[54] = 1537002063
    K[55] = 1747873779
    K[56] = 1955562222
    K[57] = 2024104815
    K[58] = -2067236844
    K[59] = -1933114872
    K[60] = -1866530822
    K[61] = -1538233109
    K[62] = -1090935817
    K[63] = -965641998

    // Hash state
    var h0 = 1779033703
    var h1 = -1150833019
    var h2 = 1013904242
    var h3 = -1521486534
    var h4 = 1359893119
    var h5 = -1694144372
    var h6 = 528734635
    var h7 = 1541459225

    // Padding
    var original_len = len(data)
    var rem = (original_len + 9) % 64
    var padding_len = 0
    if (rem != 0) {
        padding_len = 64 - rem
    }

    var total_len = original_len + 1 + padding_len + 8
    var padded = alloc byte[total_len]

    var i = 0
    while i < original_len {
        padded[i] = data[i]
        i = i + 1
    }

    padded[original_len] = 128

    i = original_len + 1
    while i < original_len + 1 + padding_len {
        padded[i] = 0
        i = i + 1
    }

    var bits_len = original_len * 8
    var len_idx = total_len - 8
    padded[len_idx] = 0
    padded[len_idx + 1] = 0
    padded[len_idx + 2] = 0
    padded[len_idx + 3] = 0
    padded[len_idx + 4] = ((bits_len >> 24) & 255)
    padded[len_idx + 5] = ((bits_len >> 16) & 255)
    padded[len_idx + 6] = ((bits_len >> 8) & 255)
    padded[len_idx + 7] = (bits_len & 255)

    // Process blocks
    var num_blocks = total_len / 64
    var block = 0
    var W = alloc i32[64]

    while block < num_blocks {
        var block_offset = block * 64

        // 1. Prepare message schedule W
        var t = 0
        while t < 16 {
            var o = block_offset + t * 4
            var b0: i32 = padded[o]
            var b1: i32 = padded[o + 1]
            var b2: i32 = padded[o + 2]
            var b3: i32 = padded[o + 3]
            W[t] = ((b0 & 255) << 24) | ((b1 & 255) << 16) | ((b2 & 255) << 8) | (b3 & 255)
            t = t + 1
        }

        while t < 64 {
            var w15 = W[t - 15]
            var s0 = rotr(w15, 7) ^ rotr(w15, 18) ^ shr_logical(w15, 3)
            var w2 = W[t - 2]
            var s1 = rotr(w2, 17) ^ rotr(w2, 19) ^ shr_logical(w2, 10)
            W[t] = W[t - 16] + s0 + W[t - 7] + s1
            t = t + 1
        }

        // 2. Initialize working registers
        var a = h0
        var b = h1
        var c = h2
        var d = h3
        var e = h4
        var f = h5
        var g = h6
        var h = h7

        // 3. Compression loop
        t = 0
        while t < 64 {
            var S1 = rotr(e, 6) ^ rotr(e, 11) ^ rotr(e, 25)
            var ch = (e & f) ^ ((~e) & g)
            var temp1 = h + S1 + ch + K[t] + W[t]

            var S0 = rotr(a, 2) ^ rotr(a, 13) ^ rotr(a, 22)
            var maj = (a & b) ^ (a & c) ^ (b & c)
            var temp2 = S0 + maj

            h = g
            g = f
            f = e
            e = d + temp1
            d = c
            c = b
            b = a
            a = temp1 + temp2

            t = t + 1
        }

        // 4. Update hash state
        h0 = h0 + a
        h1 = h1 + b
        h2 = h2 + c
        h3 = h3 + d
        h4 = h4 + e
        h5 = h5 + f
        h6 = h6 + g
        h7 = h7 + h

        block = block + 1
    }

    // Build digest
    var digest = alloc byte[32]

    digest[0] = ((h0 >> 24) & 255)
    digest[1] = ((h0 >> 16) & 255)
    digest[2] = ((h0 >> 8) & 255)
    digest[3] = (h0 & 255)
    digest[4] = ((h1 >> 24) & 255)
    digest[5] = ((h1 >> 16) & 255)
    digest[6] = ((h1 >> 8) & 255)
    digest[7] = (h1 & 255)
    digest[8] = ((h2 >> 24) & 255)
    digest[9] = ((h2 >> 16) & 255)
    digest[10] = ((h2 >> 8) & 255)
    digest[11] = (h2 & 255)
    digest[12] = ((h3 >> 24) & 255)
    digest[13] = ((h3 >> 16) & 255)
    digest[14] = ((h3 >> 8) & 255)
    digest[15] = (h3 & 255)
    digest[16] = ((h4 >> 24) & 255)
    digest[17] = ((h4 >> 16) & 255)
    digest[18] = ((h4 >> 8) & 255)
    digest[19] = (h4 & 255)
    digest[20] = ((h5 >> 24) & 255)
    digest[21] = ((h5 >> 16) & 255)
    digest[22] = ((h5 >> 8) & 255)
    digest[23] = (h5 & 255)
    digest[24] = ((h6 >> 24) & 255)
    digest[25] = ((h6 >> 16) & 255)
    digest[26] = ((h6 >> 8) & 255)
    digest[27] = (h6 & 255)
    digest[28] = ((h7 >> 24) & 255)
    digest[29] = ((h7 >> 16) & 255)
    digest[30] = ((h7 >> 8) & 255)
    digest[31] = (h7 & 255)

    return digest
}

pub fn Sha256String(s: #str) str {
    var data = StringToBytes(s)
    var hash = Sha256(data)
    return EncodeHex(hash)
}

extern fn nr_crypto_rand_bytes(p: ptr, size: i32) bool

pub fn RandBytes(size: i32) byte[0] {
    var alloc_size = size
    if (alloc_size < 0) {
        alloc_size = 0
    }
    var b = alloc byte[alloc_size]
    if (alloc_size > 0) {
        var ok = nr_crypto_rand_bytes(b, alloc_size)
        if (!ok) {
            panic("crypto: failed to generate secure random bytes")
        }
    }
    return b
}

pub fn RandInt() i32 {
    var b = alloc byte[4]
    var ok = nr_crypto_rand_bytes(b, 4)
    if (!ok) {
        panic("crypto: failed to generate secure random bytes")
    }
    var b0: i32 = b[0]
    var b1: i32 = b[1]
    var b2: i32 = b[2]
    var b3: i32 = b[3]
    return (b0 & 255) | ((b1 & 255) << 8) | ((b2 & 255) << 16) | ((b3 & 255) << 24)
}

pub fn RandIntRange(min: i32, max: i32) i32 {
    if (min >= max) {
        return min
    }
    var r = RandInt()
    if (r < 0) {
        r = -r
    }
    return min + (r % (max - min))
}
