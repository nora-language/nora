package main

import "io"

fn main() {
    var title = "Nora Interpolated String"
    var content = "This is a demonstration of multiline string support with interpolation!"
    
    var html = `<html>
    <head>
        <title>${title}</title>
    </head>
    <body>
        <h1>Hello from Nora!</h1>
        <p>${content}</p>
    </body>
</html>`

    io.PrintLn("Generated HTML:")
    io.PrintLn(html)
}
