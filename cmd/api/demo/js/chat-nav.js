'use strict';

/* Navigasi dari kartu produk di area chat */
const ChatNav = {
  toProductDetail(productId) {
    const id = parseInt(productId, 10);
    if (!id) return;
    switchTab('shop');
    Shop.openDetail(id);
  },
  toProductEdit(productId) {
    const id = parseInt(productId, 10);
    if (!id) return;
    Seller.switchView('products');
    SellerProducts.openModal(id);
  },
};
