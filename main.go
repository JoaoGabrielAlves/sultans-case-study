package main

import (
	"fmt"
	"log"

	"sultans-case-study/pkg/shopify"
	"sultans-case-study/tasks/task1"
	"sultans-case-study/tasks/task2"
)

func main() {
	fmt.Println("Shopify Interview Tasks")
	fmt.Println("=======================")

	client, err := shopify.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Shopify client: %v", err)
	}

	fmt.Println("\nTask 1: Generating customer leaderboard...")
	if err := task1.GenerateCustomerLeaderboard(client); err != nil {
		log.Printf("Task 1 failed: %v", err)
	} else {
		fmt.Println("Customer leaderboard exported successfully")
	}

	fmt.Println("\nTask 2: Capture the flag...")
	if err := task2.CaptureTheFlag(client); err != nil {
		log.Printf("Task 2 failed: %v", err)
	} else {
		fmt.Println("Flag captured successfully")
	}

	fmt.Println("\nAll tasks completed")
} 