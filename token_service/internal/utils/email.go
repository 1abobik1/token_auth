package utils

import (
    "fmt"
    "github.com/google/uuid"
)

func SendIPChangeWarningEmail(userID uuid.UUID, oldIP, newIP string) {
    fmt.Printf("[Mock Email] IP changed for %s: %s -> %s\n",
        userID, oldIP, newIP)
}
