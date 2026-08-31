CREATE TABLE "product_images" (
    "id" UUID DEFAULT gen_random_uuid(),
    "product_id" UUID NOT NULL,
    "image_url" VARCHAR(255) UNIQUE NOT NULL,
    "is_default" BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY ("id")
);

CREATE INDEX "product_id_index" ON "product_images" ("product_id");
CREATE UNIQUE INDEX "unique_default_image" 
ON "product_images" ("product_id") 
WHERE "is_default" = TRUE;

CREATE OR REPLACE FUNCTION check_product_image_limit()
RETURNS TRIGGER AS $$
BEGIN
    IF (SELECT COUNT(*) FROM "product_images" WHERE "product_id" = NEW."product_id") > 6 THEN
        RAISE EXCEPTION 'Limit reached: A product can have a maximum of 5 images.';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER "trg_product_image_limit"
BEFORE INSERT ON "product_images"
FOR EACH ROW
EXECUTE FUNCTION check_product_image_limit();
