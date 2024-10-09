package models

import (
	"net/mail"
)

func (order *Order) IsValidOrder() bool {
	if order.OrderUID == "" || order.TrackNumber == "" || order.Entry == "" {
		return false
	}

	if !isValidDelivery(&order.Delivery) {
		return false
	}

	if !isValidPayment(&order.Payment) {
		return false
	}

	if len(order.Items) == 0 {
		return false
	}
	for _, item := range order.Items {
		if !isValidItem(&item) {
			return false
		}
	}

	if order.Locale == "" || order.CustomerID == "" || order.DeliveryService == "" {
		return false
	}

	if order.DateCreated.IsZero() {
		return false
	}

	return true
}

func isValidDelivery(delivery *Delivery) bool {
	if delivery.Name == "" || delivery.Phone == "" || delivery.Zip == "" ||
		delivery.City == "" || delivery.Address == "" || delivery.Region == "" {
		return false
	}

	_, err := mail.ParseAddress(delivery.Email)
	return err == nil
}

func isValidPayment(payment *Payment) bool {
	if payment.Transaction == "" || payment.RequestID == "" ||
		payment.Currency == "" || payment.Provider == "" {
		return false
	}

	if payment.Amount <= 0 || payment.PaymentDt <= 0 ||
		payment.DeliveryCost < 0 || payment.GoodsTotal <= 0 {
		return false
	}

	return true
}

func isValidItem(item *Item) bool {
	if item.ChrtID <= 0 || item.TrackNumber == "" || item.Price <= 0 ||
		item.Rid == "" || item.Name == "" || item.NmID <= 0 || item.Brand == "" {
		return false
	}

	return true
}
