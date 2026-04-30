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

