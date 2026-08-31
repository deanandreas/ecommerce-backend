-- CART AND ORDER ITEM RELATIONS (Drop in reverse)
ALTER TABLE "order_items" DROP CONSTRAINT "order_items_product_id_foreign";
ALTER TABLE "order_items" DROP CONSTRAINT "order_items_order_id_foreign";
ALTER TABLE "cart_items" DROP CONSTRAINT "cart_items_product_id_foreign";
ALTER TABLE "cart_items" DROP CONSTRAINT "cart_items_cart_id_foreign";
ALTER TABLE "carts" DROP CONSTRAINT "carts_user_id_foreign";

-- PRODUCT RELATIONS
ALTER TABLE "reviews" DROP CONSTRAINT "reviews_product_id_foreign";
ALTER TABLE "reviews" DROP CONSTRAINT "reviews_user_id_foreign";
ALTER TABLE "product_images" DROP CONSTRAINT "product_images_product_id_foreign";
ALTER TABLE "products" DROP CONSTRAINT "products_category_id_foreign";
ALTER TABLE "products" DROP CONSTRAINT "products_user_id_foreign";

-- USER AND ADDRESS RELATIONS
ALTER TABLE "payments" DROP CONSTRAINT "payments_order_id_foreign";
ALTER TABLE "payments" DROP CONSTRAINT "payments_user_id_foreign";
ALTER TABLE "orders" DROP CONSTRAINT "orders_address_id_foreign";
ALTER TABLE "orders" DROP CONSTRAINT "orders_user_id_foreign";
ALTER TABLE "addresses" DROP CONSTRAINT "addresses_user_id_foreign";
