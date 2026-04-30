package controller

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yourname/pos-koperasi/internal/middleware"
	"github.com/yourname/pos-koperasi/internal/modules/product/contract"
	"github.com/yourname/pos-koperasi/pkg/response"
)

type Controller struct {
	svc contract.ProductService
}

func New(svc contract.ProductService) *Controller {
	return &Controller{svc: svc}
}

// GET /api/v1/products/barcode/{barcode}
func (c *Controller) GetByBarcode(w http.ResponseWriter, r *http.Request) {
	barcode := chi.URLParam(r, "barcode")
	if barcode == "" {
		response.BadRequest(w, "Barcode diperlukan")
		return
	}

	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	product, err := c.svc.GetByBarcode(barcode, claims.BranchID)
	if err != nil {
		response.NotFound(w, "Produk tidak ditemukan")
		return
	}

	if product.Stock <= 0 {
		response.JSON(w, http.StatusOK, false, "Stok produk habis", product)
		return
	}

	response.Success(w, "Produk ditemukan", product)
}

// GET /api/v1/products/low-stock
func (c *Controller) GetLowStock(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	products, err := c.svc.ListLowStock(claims.BranchID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, "Low stock products", products)
}

