package midtrans

import (
	"fmt"
	"math"
	"strings"
)

// ProductLineInput satu baris keranjang untuk item_details Midtrans (harga per unit IDR bulat).
type ProductLineInput struct {
	ProductID int
	Name      string
	UnitPrice float64
	Quantity  int
}

// CustomerInput identitas pemesan untuk customer_details.
type CustomerInput struct {
	FullName string
	Phone    string
	Email    string
}

// AddressInput alamat penagihan/pengiriman (opsional).
type AddressInput struct {
	RecipientName string
	Phone         string
	AddressLine   string
	City          string
	PostalCode    string
}

// ChargeExtras opsional untuk /v2/charge — diisi lewat NewChargeExtrasFromCart.
type ChargeExtras struct {
	oc optionalChargeFields
}

type optionalChargeFields struct {
	CustomerDetails *customerDetailsJSON `json:"customer_details,omitempty"`
	ItemDetails     []itemDetailJSON     `json:"item_details,omitempty"`
}

type customerDetailsJSON struct {
	FirstName       string             `json:"first_name,omitempty"`
	LastName        string             `json:"last_name,omitempty"`
	Email           string             `json:"email,omitempty"`
	Phone           string             `json:"phone,omitempty"`
	BillingAddress  *addressDetailJSON `json:"billing_address,omitempty"`
	ShippingAddress *addressDetailJSON `json:"shipping_address,omitempty"`
}

type addressDetailJSON struct {
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Address     string `json:"address,omitempty"`
	City        string `json:"city,omitempty"`
	PostalCode  string `json:"postal_code,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
}

type itemDetailJSON struct {
	ID       string `json:"id"`
	Price    int64  `json:"price"`
	Quantity int    `json:"quantity"`
	Name     string `json:"name"`
}

// NewChargeExtrasFromCart menyusun customer_details + item_details.
// grossAmount harus konsisten dengan baris produk + diskon voucher + PPN (sesuai logika bisnis Anda).
func NewChargeExtrasFromCart(
	grossAmount float64,
	lines []ProductLineInput,
	voucherDiscount float64,
	tax float64,
	customer CustomerInput,
	billing, shipping *AddressInput,
) (*ChargeExtras, error) {
	gross := int64(math.Round(grossAmount))
	if gross <= 0 {
		return nil, fmt.Errorf("gross_amount tidak valid")
	}

	var items []itemDetailJSON
	var productSum int64

	for _, ln := range lines {
		name := strings.TrimSpace(ln.Name)
		if name == "" {
			name = fmt.Sprintf("Produk #%d", ln.ProductID)
		}
		if len(name) > 120 {
			name = name[:120]
		}
		unit := int64(math.Round(ln.UnitPrice))
		if ln.Quantity < 1 {
			continue
		}
		id := fmt.Sprintf("P%d", ln.ProductID)
		items = append(items, itemDetailJSON{
			ID: id, Price: unit, Quantity: ln.Quantity, Name: name,
		})
		productSum += unit * int64(ln.Quantity)
	}

	vd := int64(math.Round(voucherDiscount))
	if vd > 0 {
		items = append(items, itemDetailJSON{
			ID: "DISCOUNT", Price: -vd, Quantity: 1, Name: "Diskon voucher",
		})
	}
	ti := int64(math.Round(tax))
	if ti > 0 {
		items = append(items, itemDetailJSON{
			ID: "TAX", Price: ti, Quantity: 1, Name: "PPN",
		})
	}

	computed := productSum - vd + ti
	diff := gross - computed
	if diff != 0 {
		items = append(items, itemDetailJSON{
			ID: "ADJ", Price: diff, Quantity: 1, Name: "Penyesuaian pembulatan",
		})
	}

	var check int64
	for _, it := range items {
		check += it.Price * int64(it.Quantity)
	}
	if check != gross {
		return nil, fmt.Errorf("item_details total %d != gross_amount %d (computed=%d)", check, gross, computed)
	}

	first, last := splitName(customer.FullName)
	email := strings.TrimSpace(customer.Email)
	if email == "" {
		if customer.Phone != "" {
			email = sanitizeEmailLocal(customer.Phone) + "@customers.pos.local"
		} else {
			email = "guest@customers.pos.local"
		}
	}

	cd := &customerDetailsJSON{
		FirstName: first,
		LastName:  last,
		Email:     email,
		Phone:     strings.TrimSpace(customer.Phone),
	}
	if billing != nil {
		cd.BillingAddress = mapAddress(billing)
	}
	if shipping != nil {
		cd.ShippingAddress = mapAddress(shipping)
	}
	// Jika hanya satu alamat tersedia, gunakan juga untuk yang lain agar dashboard terisi.
	if cd.BillingAddress == nil && cd.ShippingAddress != nil {
		cd.BillingAddress = cd.ShippingAddress
	}
	if cd.ShippingAddress == nil && cd.BillingAddress != nil {
		cd.ShippingAddress = cd.BillingAddress
	}

	return &ChargeExtras{oc: optionalChargeFields{
		CustomerDetails: cd,
		ItemDetails:     items,
	}}, nil
}

func mapAddress(a *AddressInput) *addressDetailJSON {
	if a == nil {
		return nil
	}
	f, l := splitName(a.RecipientName)
	return &addressDetailJSON{
		FirstName:   f,
		LastName:    l,
		Phone:       strings.TrimSpace(a.Phone),
		Address:     strings.TrimSpace(a.AddressLine),
		City:        strings.TrimSpace(a.City),
		PostalCode:  strings.TrimSpace(a.PostalCode),
		CountryCode: "IDN",
	}
}

func splitName(full string) (first, last string) {
	full = strings.TrimSpace(full)
	if full == "" {
		return "Pelanggan", ""
	}
	parts := strings.Fields(full)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func sanitizeEmailLocal(phone string) string {
	s := strings.TrimSpace(phone)
	s = strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return '-'
	}, s)
	if s == "" {
		return "no-phone"
	}
	return s
}
