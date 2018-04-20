//A sample program written to gain familiarity with Go and its
//concurrency mechanisms. Takes a number of coin flips, an interval
//between flips, and the bias of the coin. Simulates the given number
//of coin flips and attempts to predict the result of the next
//based on observed results.

package main

import (
    "time"
    "fmt"
    "math/rand"
    "math"
    "os"
    "strconv"
)

//Returns a channel which will give the coin flip results
func CoinFlipper(flipInterval time.Duration, bias float64) chan int {
    c := make(chan int)

    go func(c chan int){
        for {
            <-time.After(flipInterval)

            //Randomly set the result of the coinflip using the bias
            flipResult := 0
            if (rand.Float64() <= bias) {
                flipResult = 1
            }

            select {
                //If c is not ready to receive a value, then terminate
                case c <- flipResult:
                default:
                    close(c)
                    return
            }
        }
    }(c)
    return c
}

func main() {
    if (len(os.Args) != 4) {
        fmt.Println("usage: ./project maxFlips flipInterval bias")
        return
    }
   
    maxFlips, err := strconv.ParseInt(os.Args[1], 10, 64)
    if (err != nil || maxFlips < 0) {
        fmt.Println("Error: maxFlips must be a valid positive integer.")
        return
    }
    flipInterval, err := strconv.ParseFloat(os.Args[2], 64)
    if (err != nil || flipInterval < 0) {
        fmt.Println("Error: flipInterval must be a valid positive float.")
        return
    }
    bias, err := strconv.ParseFloat(os.Args[3], 64)
    if (err != nil || bias < 0 || bias > 1) {
        fmt.Println("Error: bias must be a valid probability (between 0 and 1).")
        return
    }

    c := CoinFlipper(time.Duration(flipInterval) * time.Second, bias)
    totalHeads := 0
    var observedBias float64 
    for i := 0; int64(i) < maxFlips; i++ {
        result := <-c

        fmt.Print(
            "I saw ",
            result,
            ". ",
        )

        totalHeads += result
        observedBias = float64(totalHeads + 1)/float64(i + 3)

        if (observedBias == .5) {
            fmt.Println("Both outcomes seem equally likely.")
        } else {
            fmt.Println(
                "I expect to see a",
                math.Round(observedBias),
                "next.",
            )
        }

    }
    
    fmt.Println(
        "I think the bias is ",
        observedBias,
        ". This is ",
        math.Abs(bias - observedBias),
        "off.",
    )
}