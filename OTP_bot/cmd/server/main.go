package main

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"otp_bot/internal/otpbot"
)

func main() {
	if err := otpbot.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
