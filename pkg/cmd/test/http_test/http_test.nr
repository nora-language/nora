package main

import "io"
import "net/http"
import "strconv"
import "time"

fn test_handler(req: #http.Request) http.Response {
    io.PrintLn("Server received Request: Method=${req.method}, URL=${req.url}")
    var ua = req.GetHeader("User-Agent")
    io.PrintLn("Server received User-Agent Header: ${ua}")
    var body = "Hello from Nora HTTP Server!"
    return http.NewResponse(200, "OK", body)
}

fn start_server(port: i32, ready_ch: #chan[bool], done_ch: #chan[bool]) {
    var server = http.Server {
        port: port
    }
    ready_ch <- true
    match server.ListenAndServeOnce(test_handler) {
        Ok(_) => {}
        Err(err) => {
            io.PrintLn("Server error: ${err}")
        }
    }
    done_ch <- true
}

fn main() i32 {
    io.PrintLn("=== Nora HTTP Standard Library Integration Test ===")

    var port = 9099
    var ready_ch = make(chan[bool], 1)
    var done_ch = make(chan[bool], 1)

    scope {

        spawn start_server(port, #ready_ch, #done_ch)

        // Wait until server is started
        var ready = <-ready_ch
        if (ready) {
            // Sleep 50ms to guarantee TCP binding is complete on OS level
            var d = time.Duration {
                nanoseconds: 50000000
            }
            time.Sleep(d)

            io.PrintLn("Client sending GET request...")
            match http.Get("127.0.0.1", port) {
                Ok(resp) => {
                    io.PrintLn("Client received Response Status: ${strconv.Itoa(resp.status_code)}")
                    io.PrintLn("Client received Response Text: ${resp.status_text}")
                    io.PrintLn("Client received Response Body: ${resp.body}")
                }
                Err(err) => {
                    io.PrintLn("Client received HTTP Dial Error: ${err}")
                }
            }

            // Wait for server fiber stack cleanup
            var _ = <-done_ch
        }

        return 0


    }
}
