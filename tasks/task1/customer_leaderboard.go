package task1

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"sultans-case-study/pkg/shopify"
)

type Customer struct {
	ID          string
	FirstName   string
	LastName    string
	Email       string
	AmountSpent float64
	Currency    string
}

func GenerateCustomerLeaderboard(client *shopify.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	customers, err := fetchFilteredCustomers(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to fetch customers: %w", err)
	}

	if len(customers) == 0 {
		return fmt.Errorf("no customers found with required tags")
	}

	topCustomers := customers
	if len(customers) > 50 {
		topCustomers = customers[:50]
	}

	return exportToCSV(topCustomers, "top_50_customers.csv")
}

func fetchFilteredCustomers(ctx context.Context, client *shopify.Client) ([]Customer, error) {
	query := `
		query GetCustomerSegmentMembers($first: Int!, $query: String!, $sortKey: String, $reverse: Boolean!) {
			customerSegmentMembers(first: $first, query: $query, sortKey: $sortKey, reverse: $reverse) {
				edges {
					node {
						id
						firstName
						lastName
						defaultEmailAddress {
							emailAddress
						}
						amountSpent {
							amount
							currencyCode
						}
					}
				}
			}
		}
	`
	
	variables := map[string]interface{}{
		"first":   50,
		"query":   "customer_tags CONTAINS 'task1' AND customer_tags CONTAINS 'level:3'",
		"sortKey": "amount_spent",
		"reverse": true,
	}

	response, err := client.GraphQL(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("GraphQL query failed: %w", err)
	}

	var result struct {
		CustomerSegmentMembers struct {
			Edges []struct {
				Node struct {
					ID        string `json:"id"`
					FirstName string `json:"firstName"`
					LastName  string `json:"lastName"`
					DefaultEmailAddress struct {
						EmailAddress string `json:"emailAddress"`
					} `json:"defaultEmailAddress"`
					AmountSpent struct {
						Amount       string `json:"amount"`
						CurrencyCode string `json:"currencyCode"`
					} `json:"amountSpent"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"customerSegmentMembers"`
	}

	if err := json.Unmarshal(response.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var customers []Customer
	for _, edge := range result.CustomerSegmentMembers.Edges {
		node := edge.Node
		
		amountSpent, err := strconv.ParseFloat(node.AmountSpent.Amount, 64)
		if err != nil {
			fmt.Printf("Warning: failed to parse amount for customer %s: %v\n", node.ID, err)
			continue
		}
		
		customers = append(customers, Customer{
			ID:          node.ID,
			FirstName:   node.FirstName,
			LastName:    node.LastName,
			Email:       node.DefaultEmailAddress.EmailAddress,
			AmountSpent: amountSpent,
			Currency:    node.AmountSpent.CurrencyCode,
		})
	}

	return customers, nil
}

func exportToCSV(customers []Customer, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Rank", "Customer ID", "First Name", "Last Name", "Email", "Amount Spent", "Currency"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for i, customer := range customers {
		record := []string{
			strconv.Itoa(i + 1),
			customer.ID,
			customer.FirstName,
			customer.LastName,
			customer.Email,
			fmt.Sprintf("%.2f", customer.AmountSpent),
			customer.Currency,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	fmt.Printf("Exported %d customers to %s\n", len(customers), filename)
	return nil
} 