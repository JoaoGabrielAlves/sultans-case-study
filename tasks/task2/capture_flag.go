package task2

import (
	"context"
	"encoding/json"
	"fmt"

	"sultans-case-study/pkg/shopify"
)

type ProductInfo struct {
	ProductID       string
	Title           string
	Handle          string
	VariantID       string
	InventoryItemID string
}

func CaptureTheFlag(client *shopify.Client) error {
	ctx := context.Background()

	flagProduct, err := findProductByHandle(ctx, client, "the-videographer-snowboard")
	if err != nil {
		return fmt.Errorf("failed to find flag product: %w", err)
	}
	fmt.Printf("Found flag product: %s\n", flagProduct.Title)

	if err := disableInventoryTracking(ctx, client, flagProduct.InventoryItemID); err != nil {
		fmt.Printf("Warning: Failed to disable inventory tracking: %v\n", err)
	}

	config := client.GetConfig()
	if err := findFlagOrders(ctx, client, flagProduct.VariantID, config.UserEmail); err != nil {
		fmt.Printf("No existing flag orders found: %v\n", err)
		
		fmt.Printf("\nTo capture the flag, place an order manually:\n")
		fmt.Printf("1. Visit: %s (password: %s)\n", config.ShopifyURL, config.StorePassword)
		fmt.Printf("2. Search for the flag product and add to cart\n")
		fmt.Printf("3. Checkout with email: %s\n", config.UserEmail)
		fmt.Printf("4. Complete the order\n")
	}

	return nil
}

func findProductByHandle(ctx context.Context, client *shopify.Client, handle string) (*ProductInfo, error) {
	query := `
		query GetProductByHandle($handle: String!) {
			product: productByHandle(handle: $handle) {
				id
				title
				handle
				variants(first: 1) {
					edges {
						node {
							id
							inventoryItem {
								id
							}
						}
					}
				}
			}
		}
	`
	
	variables := map[string]interface{}{
		"handle": handle,
	}

	response, err := client.GraphQL(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("GraphQL query failed: %w", err)
	}

	var result struct {
		Product struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			Handle   string `json:"handle"`
			Variants struct {
				Edges []struct {
					Node struct {
						ID            string `json:"id"`
						InventoryItem struct {
							ID string `json:"id"`
						} `json:"inventoryItem"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"variants"`
		} `json:"product"`
	}

	if err := json.Unmarshal(response.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Product.ID == "" {
		return nil, fmt.Errorf("no product found with handle: %s", handle)
	}

	if len(result.Product.Variants.Edges) == 0 {
		return nil, fmt.Errorf("no variants found for product")
	}

	variant := result.Product.Variants.Edges[0].Node

	return &ProductInfo{
		ProductID:       result.Product.ID,
		Title:           result.Product.Title,
		Handle:          result.Product.Handle,
		VariantID:       variant.ID,
		InventoryItemID: variant.InventoryItem.ID,
	}, nil
}

func disableInventoryTracking(ctx context.Context, client *shopify.Client, inventoryItemID string) error {
	mutation := `
		mutation UpdateInventoryItem($id: ID!, $input: InventoryItemInput!) {
			inventoryItemUpdate(id: $id, input: $input) {
				inventoryItem {
					id
					tracked
				}
				userErrors {
					field
					message
				}
			}
		}
	`
	
	variables := map[string]interface{}{
		"id": inventoryItemID,
		"input": map[string]interface{}{
			"tracked": false,
		},
	}

	response, err := client.GraphQL(ctx, mutation, variables)
	if err != nil {
		return fmt.Errorf("GraphQL mutation failed: %w", err)
	}

	var result struct {
		InventoryItemUpdate struct {
			UserErrors []struct {
				Field   string `json:"field"`
				Message string `json:"message"`
			} `json:"userErrors"`
		} `json:"inventoryItemUpdate"`
	}

	if err := json.Unmarshal(response.Data, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.InventoryItemUpdate.UserErrors) > 0 {
		firstError := result.InventoryItemUpdate.UserErrors[0]
		return fmt.Errorf("failed to update inventory tracking: %s", firstError.Message)
	}

	fmt.Printf("Inventory tracking disabled\n")
	return nil
}

func findFlagOrders(ctx context.Context, client *shopify.Client, variantID, userEmail string) error {
	query := `
		query GetOrdersByEmail($first: Int!, $query: String!) {
			orders(first: $first, query: $query) {
				edges {
					node {
						id
						name
						email
						lineItems(first: 10) {
							edges {
								node {
									variant {
										id
									}
								}
							}
						}
					}
				}
			}
		}
	`
	
	variables := map[string]interface{}{
		"first": 10,
		"query": fmt.Sprintf("email:%s", userEmail),
	}

	response, err := client.GraphQL(ctx, query, variables)
	if err != nil {
		return fmt.Errorf("GraphQL query failed: %w", err)
	}

	var result struct {
		Orders struct {
			Edges []struct {
				Node struct {
					ID        string `json:"id"`
					Name      string `json:"name"`
					Email     string `json:"email"`
					LineItems struct {
						Edges []struct {
							Node struct {
								Variant struct {
									ID string `json:"id"`
								} `json:"variant"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"lineItems"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"orders"`
	}

	if err := json.Unmarshal(response.Data, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Orders.Edges) == 0 {
		return fmt.Errorf("no orders found for email: %s", userEmail)
	}

	for _, edge := range result.Orders.Edges {
		orderNode := edge.Node
		
		for _, itemEdge := range orderNode.LineItems.Edges {
			item := itemEdge.Node
			
			if item.Variant.ID == variantID {
				fmt.Printf("FLAG CAPTURED! Found order %s with flag product!\n", orderNode.Name)
				return nil
			}
		}
	}

	return fmt.Errorf("no orders found containing the flag product")
}
