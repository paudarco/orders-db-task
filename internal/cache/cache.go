package cache

import (
	"sync"

	"github.com/paudarco/orders-db-task/internal/models"
)

type Cache struct {
	mu     sync.RWMutex
	orders map[string]*models.Order
}

func NewCache() *Cache {
	return &Cache{
		orders: make(map[string]*models.Order),
	}
}

func (c *Cache) Set(order *models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders[order.OrderUID] = order
}

func (c *Cache) Get(orderUID string) (*models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, ok := c.orders[orderUID]
	return order, ok
}

func (c *Cache) LoadFromDB(orders []*models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, order := range orders {
		c.orders[order.OrderUID] = order
	}
}
