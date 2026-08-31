-- USER AND ADDRESS RELATIONS
ALTER TABLE "addresses" ADD CONSTRAINT "addresses_user_id_foreign" FOREIGN KEY (
    "user_id"
) REFERENCES "users" ("id");
ALTER TABLE "orders" ADD CONSTRAINT "orders_user_id_foreign" FOREIGN KEY (
    "user_id"
) REFERENCES "users" ("id");
ALTER TABLE "orders" ADD CONSTRAINT "orders_address_id_foreign" FOREIGN KEY (
    "address_id"
) REFERENCES "addresses" ("id");
ALTER TABLE "payments" ADD CONSTRAINT "payments_user_id_foreign" FOREIGN KEY (
    "user_id"
) REFERENCES "users" ("id");
ALTER TABLE "payments" ADD CONSTRAINT "payments_order_id_foreign" FOREIGN KEY (
    "order_id"
) REFERENCES "orders" ("id") ON DELETE RESTRICT;

-- PRODUCT RELATIONS
ALTER TABLE "products" ADD CONSTRAINT "products_user_id_foreign" FOREIGN KEY (
    "user_id"
) REFERENCES "users" ("id") ON DELETE RESTRICT;
ALTER TABLE "products" ADD CONSTRAINT "products_category_id_foreign" FOREIGN KEY (
    "category_id"
) REFERENCES "categories" ("id") ON DELETE CASCADE;
ALTER TABLE "product_images" ADD CONSTRAINT "product_images_product_id_foreign" FOREIGN KEY (
    "product_id"
) REFERENCES "products" ("id") ON DELETE CASCADE;
ALTER TABLE "reviews" ADD CONSTRAINT "reviews_user_id_foreign" FOREIGN KEY (
    "user_id"
) REFERENCES "users" ("id") ON DELETE CASCADE;
ALTER TABLE "reviews" ADD CONSTRAINT "reviews_product_id_foreign" FOREIGN KEY (
    "product_id"
) REFERENCES "products" ("id") ON DELETE CASCADE;

-- CART AND ORDER ITEM RELATIONS
ALTER TABLE "carts" ADD CONSTRAINT "carts_user_id_foreign" FOREIGN KEY (
    "user_id"
) REFERENCES "users" ("id") ON DELETE CASCADE;
ALTER TABLE "cart_items" ADD CONSTRAINT "cart_items_cart_id_foreign" FOREIGN KEY (
    "cart_id"
) REFERENCES "carts" ("id") ON DELETE CASCADE;
ALTER TABLE "cart_items" ADD CONSTRAINT "cart_items_product_id_foreign" FOREIGN KEY (
    "product_id"
) REFERENCES "products" ("id");
ALTER TABLE "order_items" ADD CONSTRAINT "order_items_order_id_foreign" FOREIGN KEY (
    "order_id"
) REFERENCES "orders" ("id") ON DELETE RESTRICT;
ALTER TABLE "order_items" ADD CONSTRAINT "order_items_product_id_foreign" FOREIGN KEY (
    "product_id"
) REFERENCES "products" ("id") ON DELETE RESTRICT;;
