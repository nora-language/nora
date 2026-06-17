package main

import "io"
import "net"

fn assert_string_eq(val: str, expected: str, name: str) {
    if (val != expected) {
        io.PrintLn("Assertion failed for: ")
        io.PrintLn(name)
        io.PrintLn("Expected: ")
        io.PrintLn(expected)
        io.PrintLn("Got: ")
        io.PrintLn(val)
        panic("Error: assertion failed")
    }
}

[unsafe]
fn run_server(ready_ch: #chan[bool]) {
    io.PrintLn("  [Server] Binding UDP socket on port 8091...")
    match net.BindUDP(8091) {
        Ok(server) => {
            io.PrintLn("  [Server] UDP Bound. Signal client...")
            ready_ch <- true

            io.PrintLn("  [Server] Waiting for UDP datagram...")
            match server.RecvFrom(1024) {
                Ok(res) => {
                    // Extract data as string
                    var data_len = len(res.data)
                    var req_str: str = alloc byte[data_len + 1]
                    var i = 0
                    while (i < data_len) {
                        req_str[i] = res.data[i]
                        i = i + 1
                    }
                    req_str[data_len] = 0

                    io.PrintLn("  [Server] Received from ${res.addr}:${res.port}: ${req_str}")

                    // Echo response back to client's port
                    var resp_msg = "UDP Echo: " + req_str
                    match server.SendStringTo(resp_msg, res.addr, res.port) {
                        Ok(sent) => {
                            io.PrintLn("  [Server] Echoed response successfully.")
                        }
                        Err(err) => {
                            panic("  [Server] SendTo failed: " + err)
                        }
                    }
                }
                Err(err) => {
                    panic("  [Server] RecvFrom failed: " + err)
                }
            }
        }
        Err(err) => {
            panic("  [Server] BindUDP failed: " + err)
        }
    }
    io.PrintLn("  [Server] Server finished successfully.")
}

[unsafe]
fn main() {
    io.PrintLn("=== Starting Nora Premium UDP Networking Test Suite ===")

    var ready_ch = make(chan[bool], 0)

    // Spawn server coroutine
    scope {
        spawn run_server(#ready_ch)

        // Wait for server to be ready
        io.PrintLn("  [Client] Waiting for server UDP signal...")
        var _is_ready = <-ready_ch
        io.PrintLn("  [Client] Signal received. Server is listening!")

        // Bind client UDP socket on port 8092
        io.PrintLn("  [Client] Binding UDP client on port 8092...")
        match net.BindUDP(8092) {
            Ok(client) => {
                // Send message to server
                var msg = "Hello Nora UDP Sockets!"
                io.PrintLn("  [Client] Sending datagram to 127.0.0.1:8091...")
                match client.SendStringTo(msg, "127.0.0.1", 8091) {
                    Ok(sent) => {
                        io.PrintLn("  [Client] Datagram sent.")
                    }
                    Err(err) => {
                        panic("  [Client] SendStringTo failed: " + err)
                    }
                }

                // Receive echoed response
                io.PrintLn("  [Client] Waiting for UDP response...")
                match client.RecvFrom(1024) {
                    Ok(res) => {
                        var data_len = len(res.data)
                        var resp_str: str = alloc byte[data_len + 1]
                        var i = 0
                        while (i < data_len) {
                            resp_str[i] = res.data[i]
                            i = i + 1
                        }
                        resp_str[data_len] = 0

                        io.PrintLn("  [Client] Received response from ${res.addr}:${res.port}: ${resp_str}")
                        assert_string_eq(resp_str, "UDP Echo: Hello Nora UDP Sockets!", "UDP response validation")
                    }
                    Err(err) => {
                        panic("  [Client] RecvFrom failed: " + err)
                    }
                }
            }
            Err(err) => {
                panic("  [Client] Client BindUDP failed: " + err)
            }
        }

        io.PrintLn("=== All Standard UDP Networking Tests Passed Successfully! ===")

    }
}

