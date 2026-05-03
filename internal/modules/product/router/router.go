package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/yourname/pos-koperasi/internal/modules/product/controller"
)

type Router struct {
	product *controller.Controller
}

func New(product *controller.Controller) *Router {
	return &Router{product: product}
}

func (rt *Router) RegisterProtected(r chi.Router) {
	r.Get("/products/barcode/{barcode}", rt.product.GetByBarcode)
	r.Get("/products/low-stock", rt.product.GetLowStock)
}

// RegisterPublic mendaftarkan endpoint produk yang tidak membutuhkan autentikasi (shop demo).
func (rt *Router) RegisterPublic(r chi.Router) {
	r.Get("/shop/products", rt.product.ListAll)
	r.Get("/shop/products/{id}", rt.product.GetByIDPublic)
}

// RegisterSeller CRUD produk untuk dashboard penjual (JWT karyawan).
func (rt *Router) RegisterSeller(r chi.Router) {
	r.Get("/seller/products", rt.product.ListForSeller)
	r.Post("/seller/products", rt.product.CreateSeller)
	r.Put("/seller/products/{id}", rt.product.UpdateSeller)
	r.Delete("/seller/products/{id}", rt.product.DeleteSeller)
}

