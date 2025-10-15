package orders

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"errandShop/internal/domain/products"
)

type CartService struct {
	db          *gorm.DB
	productRepo *products.Repository
}

func NewCartService(db *gorm.DB, productRepo *products.Repository) *CartService {
	return &CartService{
		db:          db,
		productRepo: productRepo,
	}
}

// GetOrCreateCart gets user's cart or creates one if it doesn't exist
func (s *CartService) GetOrCreateCart(userID uuid.UUID) (*Cart, error) {
	var cart Cart
	err := s.db.Preload("Items.Product").Where("user_id = ?", userID).First(&cart).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new cart
			cart = Cart{
				UserID: userID,
			}
			if err := s.db.Create(&cart).Error; err != nil {
				return nil, fmt.Errorf("failed to create cart: %w", err)
			}
			return &cart, nil
		}
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}
	return &cart, nil
}

// AddToCart adds an item to the user's cart
func (s *CartService) AddToCart(userID uuid.UUID, req AddToCartRequest) (*Cart, error) {
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	// Check if item already exists in cart
	var existingItem CartItem
	err = s.db.Where("cart_id = ? AND product_id = ?", cart.ID, req.ProductID).First(&existingItem).Error
	if err == nil {
		// Update quantity
		existingItem.Quantity += req.Quantity
		if err := s.db.Save(&existingItem).Error; err != nil {
			return nil, fmt.Errorf("failed to update cart item: %w", err)
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Add new item
		newItem := CartItem{
			CartID:    cart.ID,
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		}
		if err := s.db.Create(&newItem).Error; err != nil {
			return nil, fmt.Errorf("failed to add cart item: %w", err)
		}
	} else {
		return nil, fmt.Errorf("failed to check existing cart item: %w", err)
	}

	// Reload cart with items
	return s.GetOrCreateCart(userID)
}

// UpdateCartItem updates the quantity of a cart item
func (s *CartService) UpdateCartItem(userID uuid.UUID, itemID uuid.UUID, req UpdateCartItemRequest) (*Cart, error) {
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	// Find the cart item
	var cartItem CartItem
	err = s.db.Where("id = ? AND cart_id = ?", itemID, cart.ID).First(&cartItem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("cart item not found")
		}
		return nil, fmt.Errorf("failed to find cart item: %w", err)
	}

	// Update quantity
	cartItem.Quantity = req.Quantity
	if err := s.db.Save(&cartItem).Error; err != nil {
		return nil, fmt.Errorf("failed to update cart item: %w", err)
	}

	// Reload cart with items
	return s.GetOrCreateCart(userID)
}

// RemoveFromCart removes an item from the cart
func (s *CartService) RemoveFromCart(userID uuid.UUID, itemID uuid.UUID) (*Cart, error) {
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	// Delete the cart item
	err = s.db.Where("id = ? AND cart_id = ?", itemID, cart.ID).Delete(&CartItem{}).Error
	if err != nil {
		return nil, fmt.Errorf("failed to remove cart item: %w", err)
	}

	// Reload cart with items
	return s.GetOrCreateCart(userID)
}

// ClearCart removes all items from the cart
func (s *CartService) ClearCart(userID uuid.UUID) error {
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return err
	}

	// Delete all cart items
	err = s.db.Where("cart_id = ?", cart.ID).Delete(&CartItem{}).Error
	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	return nil
}

// GetCart gets the user's cart
func (s *CartService) GetCart(userID uuid.UUID) (*Cart, error) {
	return s.GetOrCreateCart(userID)
}

// ValidateCartForCheckout validates that cart items are still available and prices are current
func (s *CartService) ValidateCartForCheckout(userID uuid.UUID) (*Cart, []string, error) {
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return nil, nil, err
	}

	if cart.IsEmpty() {
		return nil, []string{"Cart is empty"}, nil
	}

	var warnings []string

	// TODO: Add validation logic for:
	// - Product availability
	// - Stock levels
	// - Price changes
	// - Product status (active/inactive)

	return cart, warnings, nil
}

// ConvertCartToOrderItems converts cart items to order items with current prices
func (s *CartService) ConvertCartToOrderItems(cart *Cart) ([]CreateOrderItemRequest, error) {
	var orderItems []CreateOrderItemRequest

	for _, item := range cart.Items {
		// Get product details to include name and SKU
		product, err := s.productRepo.GetByID(context.Background(), item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product %s: %w", item.ProductID, err)
		}

		orderItems = append(orderItems, CreateOrderItemRequest{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Name:      product.Name,
			SKU:       product.SKU,
		})
	}

	return orderItems, nil
}