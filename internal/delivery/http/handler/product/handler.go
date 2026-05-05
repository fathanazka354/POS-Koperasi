package product

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fathanazka354/pos-koperasi/internal/middleware"
	"github.com/fathanazka354/pos-koperasi/internal/usecase/product"
	"github.com/fathanazka354/pos-koperasi/pkg/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	uc product.Usecase
}

func New(uc product.Usecase) *Handler {
	return &Handler{uc: uc}
}

// GET /api/v1/products/barcode/{barcode}
func (h *Handler) GetByBarcode(w http.ResponseWriter, r *http.Request) {
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
	p, err := h.uc.GetByBarcode(barcode, claims.BranchID)
	if err != nil {
		response.NotFound(w, "Produk tidak ditemukan")
		return
	}
	if p.Stock <= 0 {
		response.JSON(w, http.StatusOK, false, "Stok produk habis", p)
		return
	}
	response.Success(w, "Produk ditemukan", p)
}

// GET /api/v1/shop/products?search=&page=
func (h *Handler) ListAll(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	products, err := h.uc.ListAll(1, search, page)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "Daftar produk", products)
}

// GET /api/v1/shop/products/{id}
func (h *Handler) GetByIDPublic(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id == 0 {
		response.BadRequest(w, "ID produk tidak valid")
		return
	}
	p, err := h.uc.GetByID(id, 1)
	if err != nil {
		response.NotFound(w, "Produk tidak ditemukan")
		return
	}
	response.Success(w, "Detail produk", p)
}

// GET /api/v1/products/low-stock
func (h *Handler) GetLowStock(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	products, err := h.uc.ListLowStock(claims.BranchID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "Low stock products", products)
}

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
func (h *Handler) ListForSeller(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	search := r.URL.Query().Get("search")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	list, err := h.uc.ListForSeller(claims.BranchID, search, page)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "Daftar produk", list)
}

// POST /api/v1/seller/products
func (h *Handler) CreateSeller(w http.ResponseWriter, r *http.Request) {
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
	out, err := h.uc.CreateForSeller(claims.BranchID, product.CreateProductInput{
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
func (h *Handler) UpdateSeller(w http.ResponseWriter, r *http.Request) {
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
	out, err := h.uc.UpdateForSeller(claims.BranchID, id, product.UpdateProductInput{
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
func (h *Handler) DeleteSeller(w http.ResponseWriter, r *http.Request) {
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
	if err := h.uc.DeleteForSeller(claims.BranchID, id); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Success(w, "Produk dinonaktifkan", map[string]int{"id": id})
}

