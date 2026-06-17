package http

import "collections"
import "io"
import "net"
import "strconv"
import "string"

pub type Header = struct {
    key: str,
    value: str
}

pub type Request = struct {
    method: str,
    url: str,
    proto: str,
    headers: @collections.Vector[Header],
    body: str
}

pub type Response = struct {
    proto: str,
    status_code: i32,
    status_text: str,
    headers: @collections.Vector[Header],
    body: str
}

pub type Server = struct {
    port: i32
}

pub fn (self: &Request) AddHeader(key: #str, value: #str) {
    self.headers.Push(Header {
        key: "" + key,
        value: "" + value
    })
}

pub fn (self: #Request) GetHeader(key: #str) str {
    var n = self.headers.Len()
    var i = 0
    while (i < n) {
        var h = self.headers.Get(i)
        var hk_lower = string.ToLower(h.key)
        var k_lower = string.ToLower(key)
        if (string.Equals(hk_lower, k_lower)) {
            return h.value
        }
        i = i + 1
    }
    return ""
}

pub fn (self: &Response) AddHeader(key: #str, value: #str) {
    self.headers.Push(Header {
        key: "" + key,
        value: "" + value
    })
}

pub fn (self: #Response) GetHeader(key: #str) str {
    var n = self.headers.Len()
    var i = 0
    while (i < n) {
        var h = self.headers.Get(i)
        var hk_lower = string.ToLower(h.key)
        var k_lower = string.ToLower(key)
        if (string.Equals(hk_lower, k_lower)) {
            return h.value
        }
        i = i + 1
    }
    return ""
}

pub fn NewResponse(status_code: i32, status_text: #str, body: #str) Response {
    var headers = collections.NewVector[Header](4)
    var resp = Response {
        proto: "HTTP/1.1",
        status_code: status_code,
        status_text: "" + status_text,
        headers: @headers,
        body: "" + body
    }
    resp.AddHeader("Content-Length", strconv.Itoa(string.Length(body)))
    resp.AddHeader("Content-Type", "text/plain")
    return resp
}

pub fn ParseRequest(req_str: #str) Request {
    var sep_idx = string.IndexOf(req_str, "\r\n\r\n")
    var header_part = ""
    var body_part = ""
    if (sep_idx >= 0) {
        header_part = string.Substring(req_str, 0, sep_idx)
        body_part = string.Substring(req_str, sep_idx + 4, string.Length(req_str))
    } else {
        header_part = "" + req_str
    }

    var lines = string.Split(header_part, "\r\n")
    var req_line = lines.Get(0)
    var parts = string.Split(req_line, " ")

    var method = "" + "GET"
    var url = "" + "/"
    var proto = "" + "HTTP/1.1"
    if (parts.Len() >= 3) {
        method = "" + parts.Get(0)
        url = "" + parts.Get(1)
        proto = "" + parts.Get(2)
    }

    var headers = collections.NewVector[Header](4)
    var i = 1
    var num_lines = lines.Len()
    while (i < num_lines) {
        var line = string.Trim(lines.Get(i))
        if (string.Length(line) > 0) {
            var colon_idx = string.IndexOf(line, ":")
            if (colon_idx > 0) {
                var key_raw = string.Substring(line, 0, colon_idx)
                var key = string.Trim(key_raw)
                var val_raw = string.Substring(line, colon_idx + 1, string.Length(line))
                var value = string.Trim(val_raw)
                headers.Push(Header {
                    key: @key,
                    value: @value
                })
            }
        }
        i = i + 1
    }

    return Request {
        method: method,
        url: url,
        proto: proto,
        headers: @headers,
        body: body_part
    }
}

pub fn ParseResponse(resp_str: #str) Response {
    var sep_idx = string.IndexOf(resp_str, "\r\n\r\n")
    var header_part = ""
    var body_part = ""
    if (sep_idx >= 0) {
        header_part = string.Substring(resp_str, 0, sep_idx)
        body_part = string.Substring(resp_str, sep_idx + 4, string.Length(resp_str))
    } else {
        header_part = "" + resp_str
    }

    var lines = string.Split(header_part, "\r\n")
    var resp_line = lines.Get(0)
    var parts = string.Split(resp_line, " ")

    var proto = "" + "HTTP/1.1"
    var status_code = 200
    var status_text = "" + "OK"
    if (parts.Len() >= 3) {
        proto = "" + parts.Get(0)
        match strconv.ParseInt(parts.Get(1), 10) {
            Ok(code) => {
                status_code = code
            }
            Err(_) => {}
        }
        status_text = "" + parts.Get(2)
    }

    var headers = collections.NewVector[Header](4)
    var i = 1
    var num_lines = lines.Len()
    while (i < num_lines) {
        var line = string.Trim(lines.Get(i))
        if (string.Length(line) > 0) {
            var colon_idx = string.IndexOf(line, ":")
            if (colon_idx > 0) {
                var key_raw = string.Substring(line, 0, colon_idx)
                var key = string.Trim(key_raw)
                var val_raw = string.Substring(line, colon_idx + 1, string.Length(line))
                var value = string.Trim(val_raw)
                headers.Push(Header {
                    key: @key,
                    value: @value
                })
            }
        }
        i = i + 1
    }

    return Response {
        proto: proto,
        status_code: status_code,
        status_text: status_text,
        headers: @headers,
        body: body_part
    }
}

pub fn FormatResponse(resp: #Response) str {
    var s = "${resp.proto} ${strconv.Itoa(resp.status_code)} ${resp.status_text}\r\n"
    var n = resp.headers.Len()
    var i = 0
    while (i < n) {
        var h = resp.headers.Get(i)
        s = "${s}${h.key}: ${h.value}\r\n"
        i = i + 1
    }
    s = "${s}\r\n${resp.body}"
    return s
}

fn handle_client(conn: @net.TCPConn, handler: #fn(#Request) Response) {
    // Read the HTTP request string up to 8192 bytes
    match conn.RecvString(8192) {
        Ok(req_str) => {
            var req = ParseRequest(req_str)
            var resp = handler(#req)
            var resp_str = FormatResponse(#resp)
            var _ = conn.SendString(resp_str)
        }
        Err(_) => {}
    }
}

pub fn (self: &Server) ListenAndServe(handler: fn(#Request) Response) Result[bool, str] {
    var listener = net.Listen(self.port)?
    scope {
        while (true) {
            var conn = listener.Accept()?
            spawn handle_client(@conn, #handler)
        }
    }
    return Ok[bool, str](true)
}

pub fn (self: &Server) ListenAndServeOnce(handler: fn(#Request) Response) Result[bool, str] {
    var listener = net.Listen(self.port)?
    var conn = listener.Accept()?
    handle_client(@conn, #handler)
    return Ok[bool, str](true)
}

pub fn Get(host: #str, port: i32) Result[Response, str] {
    var conn = net.Dial(host, port)?
    var req_str = "GET / HTTP/1.1\r\nHost: ${host}:${strconv.Itoa(port)}\r\nConnection: close\r\n\r\n"
    conn.SendString(req_str)?

    // Read up to 65536 bytes of response
    var resp_str = conn.RecvString(65536)?
    var resp = ParseResponse(resp_str)
    return Ok[Response, str](resp)
}
