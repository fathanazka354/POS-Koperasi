package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

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

// GET /api/v1/shop/products?search=&page=
func (c *Controller) ListAll(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	// Gunakan branch_id=1 sebagai default untuk demo shop
	products, err := c.svc.ListAll(1, search, page)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "Daftar produk", products)
}

// GET /api/v1/shop/products/{id}
func (c *Controller) GetByIDPublic(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id == 0 {
		response.BadRequest(w, "ID produk tidak valid")
		return
	}
	product, err := c.svc.GetByID(id, 1)
	if err != nil {
		response.NotFound(w, "Produk tidak ditemukan")
		return
	}
	response.Success(w, "Detail produk", product)
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

// sellerProductReq body umum create/update produk (karyawan).
type sellerProductReq struct {
	CategoryID int     `json:"category_id"`
	SupplierID int     `json:"supplier_id"`
	Barcode    string  `json:"barcode"`
	Name       string  `json:"name"`
	Unit       string  `json:"unit"`
	BuyPrice   float64 `json:"buy_price"`
	SellPrice  float64 `json:"sell_price"`
	MinStock   int     `json:"min_stock"`
	Stock      int     `json:"stock"`
}

// GET /api/v1/seller/products — JWT karyawan
func (c *Controller) ListForSeller(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	search := r.URL.Query().Get("search")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	list, err := c.svc.ListForSeller(claims.BranchID, search, page)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "Daftar produk", list)
}

// POST /api/v1/seller/products
func (c *Controller) CreateSeller(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	var req sellerProductReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON")
		return
	}
	out, err := c.svc.CreateForSeller(claims.BranchID, contract.CreateProductInput{
		CategoryID: req.CategoryID,
		SupplierID: req.SupplierID,
		Barcode:    req.Barcode,
		Name:       req.Name,
		Unit:       req.Unit,
		BuyPrice:   req.BuyPrice,
		SellPrice:  req.SellPrice,
		MinStock:   req.MinStock,
		Stock:      req.Stock,
	})
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Created(w, "Produk dibuat", out)
}

// PUT /api/v1/seller/products/{id}
func (c *Controller) UpdateSeller(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		response.BadRequest(w, "ID tidak valid")
		return
	}
	var req sellerProductReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON")
		return
	}
	out, err := c.svc.UpdateForSeller(claims.BranchID, id, contract.UpdateProductInput{
		CategoryID: req.CategoryID,
		SupplierID: req.SupplierID,
		Barcode:    req.Barcode,
		Name:       req.Name,
		Unit:       req.Unit,
		BuyPrice:   req.BuyPrice,
		SellPrice:  req.SellPrice,
		MinStock:   req.MinStock,
		Stock:      req.Stock,
	})
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Success(w, "Produk diperbarui", out)
}

// DELETE /api/v1/seller/products/{id}
func (c *Controller) DeleteSeller(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		response.BadRequest(w, "ID tidak valid")
		return
	}
	if err := c.svc.DeleteForSeller(claims.BranchID, id); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Success(w, "Produk dinonaktifkan", map[string]int{"id": id})
}

