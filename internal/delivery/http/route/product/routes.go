package product

import (
	producthandler "github.com/fathanazka354/pos-koperasi/internal/delivery/http/handler/product"
	"github.com/go-chi/chi/v5"
)

type Routes struct {
	h *producthandler.Handler
}

func New(h *producthandler.Handler) *Routes {
	return &Routes{h: h}
}

func (rt *Routes) RegisterProtected(r chi.Router) {
	r.Get("/products/barcode/{barcode}", rt.h.GetByBarcode)
	r.Get("/products/low-stock", rt.h.GetLowStock)
}

func (rt *Routes) RegisterPublic(r chi.Router) {
	r.Get("/shop/products", rt.h.ListAll)
	r.Get("/shop/products/{id}", rt.h.GetByIDPublic)
}

func (rt *Routes) RegisterSeller(r chi.Router) {
	r.Get("/seller/products", rt.h.ListForSeller)
	r.Post("/seller/products", rt.h.CreateSeller)
	r.Put("/seller/products/{id}", rt.h.UpdateSeller)
	r.Delete("/seller/products/{id}", rt.h.DeleteSeller)
}

// RegisterSupervisor mendaftarkan endpoint khusus supervisor/admin.
// Endpoint tetap sama seperti legacy untuk menjaga kompatibilitas.
func (rt *Routes) RegisterSupervisor(r chi.Router) {
	r.Get("/products/low-stock/detail", rt.h.GetLowStock)
}

