package net

// FFI socket helpers
extern fn nr_net_dial(addr: #str, port: i32) i32

extern fn nr_net_listen(port: i32) i32

extern fn nr_net_accept(listener_fd: i32) i32

extern fn nr_net_send(socket_fd: i32, data: #byte[0], len: i32) i32

extern fn nr_net_recv(socket_fd: i32, max_len: i32) byte[0]

extern fn nr_net_close(socket_fd: i32)

extern fn strlen(s: #str) i32

extern fn nr_net_udp_socket() i32

extern fn nr_net_udp_bind(socket_fd: i32, port: i32) i32

extern fn nr_net_udp_sendto(socket_fd: i32, data: #byte[0], len: i32, addr: #str, port: i32) i32

extern fn nr_net_udp_recvfrom(socket_fd: i32, max_len: i32, out_addr: #byte[0], out_port: &i32) byte[0]

// TCPConn represents an active network connection.
pub type TCPConn = struct {
    fd: i32
}

// TCPListener represents a server socket listening for connections.
pub type TCPListener = struct {
    fd: i32
}

// UDPRecvResult represents the packet received via UDP including sender information.
pub type UDPRecvResult = struct {
    data: byte[0],
    addr: str,
    port: i32
}

// UDPSocket represents a UDP socket.
pub type UDPSocket = struct {
    fd: i32
}

// Dial connects to a remote address on the specified port.
pub fn Dial(addr: #str, port: i32) Result[TCPConn, str] {
    var fd = nr_net_dial(addr, port)
    if (fd < 0) {
        if (fd == -1) {
            return Err[TCPConn, str]("failed to create socket")
        }
        if (fd == -2) {
            return Err[TCPConn, str]("failed to resolve host address")
        }
        if (fd == -3) {
            return Err[TCPConn, str]("failed to connect to host")
        }
        return Err[TCPConn, str]("unknown connection error")
    }
    return Ok[TCPConn, str](TCPConn {
        fd: fd
    })
}

// Listen starts a local TCP server on the specified port.
pub fn Listen(port: i32) Result[TCPListener, str] {
    var fd = nr_net_listen(port)
    if (fd < 0) {
        if (fd == -1) {
            return Err[TCPListener, str]("failed to create server socket")
        }
        if (fd == -2) {
            return Err[TCPListener, str]("failed to bind server socket to port")
        }
        if (fd == -3) {
            return Err[TCPListener, str]("failed to listen on server socket")
        }
        return Err[TCPListener, str]("unknown server error")
    }
    return Ok[TCPListener, str](TCPListener {
        fd: fd
    })
}

// drop automatically closes the connection socket descriptor when going out of scope.
pub fn (self: &TCPConn) drop() {
    if (self.fd >= 0) {
        nr_net_close(self.fd)
    }
}

// drop automatically closes the listener socket descriptor when going out of scope.
pub fn (self: &TCPListener) drop() {
    if (self.fd >= 0) {
        nr_net_close(self.fd)
    }
}

// Accept blocks and accepts an incoming connection from a client.
pub fn (self: &TCPListener) Accept() Result[TCPConn, str] {
    var client_fd = nr_net_accept(self.fd)
    if (client_fd < 0) {
        return Err[TCPConn, str]("failed to accept connection")
    }
    return Ok[TCPConn, str](TCPConn {
        fd: client_fd
    })
}

// Send sends binary data across the connection.
pub fn (self: &TCPConn) Send(data: #byte[0]) Result[i32, str] {
    var n = nr_net_send(self.fd, data, len(data))
    if (n < 0) {
        return Err[i32, str]("failed to send data")
    }
    return Ok[i32, str](n)
}

// SendString sends a string across the connection.
pub fn (self: &TCPConn) SendString(s: #str) Result[i32, str] {
    // Convert string to bytes
    var len_s = strlen(s)
    var b = alloc byte[len_s]
    var i = 0
    while (i < len_s) {
        b[i] = s[i]
        i = i + 1
    }
    var res = self.Send(b)
    return res
}

// Recv receives up to max_len binary bytes from the connection.
pub fn (self: &TCPConn) Recv(max_len: i32) Result[byte[0], str] {
    var buf = nr_net_recv(self.fd, max_len)
    if (len(buf) == 0) {
        return Err[byte[0], str]("connection closed or receive error")
    }
    return Ok[byte[0], str](buf)
}

// RecvString receives up to max_len bytes and converts them into a string.
pub fn (self: &TCPConn) RecvString(max_len: i32) Result[str, str] {
    match self.Recv(max_len) {
        Ok(buf) => {
            var len_buf = len(buf)
            var s = alloc byte[len_buf + 1]
            var i = 0
            while (i < len_buf) {
                s[i] = buf[i]
                i = i + 1
            }
            s[len_buf] = 0
            return Ok[str, str](s)
        }
        Err(err) => {
            return Err[str, str](err)
        }
    }
}

// BindUDP binds a UDP socket to the specified local port.
pub fn BindUDP(port: i32) Result[UDPSocket, str] {
    var fd = nr_net_udp_socket()
    if (fd < 0) {
        return Err[UDPSocket, str]("failed to create UDP socket")
    }
    var res = nr_net_udp_bind(fd, port)
    if (res < 0) {
        nr_net_close(fd)
        return Err[UDPSocket, str]("failed to bind UDP socket to port")
    }
    return Ok[UDPSocket, str](UDPSocket {
        fd: fd
    })
}

// drop automatically closes the UDP socket descriptor when going out of scope.
pub fn (self: &UDPSocket) drop() {
    if (self.fd >= 0) {
        nr_net_close(self.fd)
    }
}

// SendTo sends binary data to a remote address and port.
pub fn (self: &UDPSocket) SendTo(data: #byte[0], addr: #str, port: i32) Result[i32, str] {
    var n = nr_net_udp_sendto(self.fd, data, len(data), addr, port)
    if (n < 0) {
        if (n == -2) {
            return Err[i32, str]("failed to resolve target address")
        }
        return Err[i32, str]("failed to send UDP packet")
    }
    return Ok[i32, str](n)
}

// SendStringTo sends a string to a remote address and port.
pub fn (self: &UDPSocket) SendStringTo(s: #str, addr: #str, port: i32) Result[i32, str] {
    // Convert string to bytes
    var len_s = strlen(s)
    var b = alloc byte[len_s]
    var i = 0
    while (i < len_s) {
        b[i] = s[i]
        i = i + 1
    }
    var res = self.SendTo(b, addr, port)
    return res
}

// RecvFrom receives up to max_len bytes from the UDP socket.
pub fn (self: &UDPSocket) RecvFrom(max_len: i32) Result[UDPRecvResult, str] {
    var out_addr_buf = alloc byte[64]
    var out_port = 0
    var buf = nr_net_udp_recvfrom(self.fd, max_len, out_addr_buf, out_port)
    if (len(buf) == 0) {
        return Err[UDPRecvResult, str]("failed to receive UDP packet")
    }

    // Convert out_addr_buf (null-terminated C-string) to Nora str
    var ip_len = 0
    while (out_addr_buf[ip_len] != 0) {
        ip_len = ip_len + 1
    }

    var ip_str = alloc byte[ip_len + 1]
    var i = 0
    while (i < ip_len) {
        ip_str[i] = out_addr_buf[i]
        i = i + 1
    }
    ip_str[ip_len] = 0

    return Ok[UDPRecvResult, str](UDPRecvResult {
        data: buf,
        addr: ip_str,
        port: out_port
    })
}
