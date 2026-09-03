package main

import (
    "fmt"
    "time"
)

func main() {
    fmt.Println("========================================")
    fmt.Println("       WELCOME TO TURBO GO 1.0!        ")
    fmt.Println("========================================")
    fmt.Println("Classic Borland Turbo Pascal / C Feel")
    fmt.Println("Running with the modern Go 1.26 toolchain!")
    fmt.Println()

    for i := 1; i <= 5; i++ {
        fmt.Printf(" [Step %d] Computing retro magic...\n", i)
        time.Sleep(100 * time.Millisecond)
    }

    fmt.Println()
    fmt.Println("Turbo Go Execution Finished successfully.")
}
