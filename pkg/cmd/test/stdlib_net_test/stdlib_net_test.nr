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

fn run_server(ready_ch: #chan[bool]) {
    io.PrintLn("  [Server] Starting listener on port 8089...")
    match net.Listen(8089) {
        Ok(listener) => {
            io.PrintLn("  [Server] Listener started. Signal client...")
            ready_ch <- true

            io.PrintLn("  [Server] Waiting for connection...")
            match listener.Accept() {
                Ok(conn) => {
                    io.PrintLn("  [Server] Accepted client connection!")

                    // Receive message
                    match conn.RecvString(1024) {
                        Ok(req) => {
                            io.PrintLn("  [Server] Received: ${req}")
                            var resp = "Echo: " + req

                            // Send response back
                            match conn.SendString(resp) {
                                Ok(sent) => {
                                    io.PrintLn("  [Server] Sent echo response")
                                }
                                Err(err) => {
                                    panic("  [Server] Send response failed: " + err)
                                }
                            }
                        }
                        Err(err) => {
                            panic("  [Server] Receive failed: " + err)
                        }
                    }
                }
                Err(err) => {
                    panic("  [Server] Accept failed: " + err)
                }
            }
        }
        Err(err) => {
            panic("  [Server] Listen failed: " + err)
        }
    }
    io.PrintLn("  [Server] Server finished successfully.")
}

fn main() {
    io.PrintLn("=== Starting Nora Premium Networking Test Suite ===")

    var ready_ch = make(chan[bool], 0)

    // Spawn server coroutine
    scope {
        spawn run_server(#ready_ch)

        // Wait for server to be ready
        io.PrintLn("  [Client] Waiting for server channel signal...")
        var _is_ready = <-ready_ch
        io.PrintLn("  [Client] Signal received. Server is listening!")

        // Connect to local server
        io.PrintLn("  [Client] Connecting to 127.0.0.1:8089...")
        match net.Dial("127.0.0.1", 8089) {
            Ok(client) => {
                io.PrintLn("  [Client] Connected to server!")

                // Send request string
                var req_msg = "Hello Nora Sockets!"
                match client.SendString(req_msg) {
                    Ok(sent) => {
                        io.PrintLn("  [Client] Sent request string")
                    }
                    Err(err) => {
                        panic("  [Client] Send failed: " + err)
                    }
                }

                // Receive echoed response
                match client.RecvString(1024) {
                    Ok(resp) => {
                        io.PrintLn("  [Client] Received response: ${resp}")
                        assert_string_eq(resp, "Echo: Hello Nora Sockets!", "Echo response verification")
                    }
                    Err(err) => {
                        panic("  [Client] Receive failed: " + err)
                    }
                }
            }
            Err(err) => {
                panic("  [Client] Dial failed: " + err)
            }
        }

        io.PrintLn("=== All Standard Networking Tests Passed Successfully! ===")

    }
}
