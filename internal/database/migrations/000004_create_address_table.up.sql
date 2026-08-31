CREATE TABLE "addresses" (
    "id" uuid DEFAULT gen_random_uuid (),
    "user_id" uuid NOT NULL,
    "street_line_1" varchar(255) NOT NULL,
    "street_line_2" varchar(255) NULL,
    "postal_code" varchar(255) NOT NULL,
    "state" varchar(255) NOT NULL,
    "city" varchar(255) NOT NULL,
    "country" varchar(255) NOT NULL,
    "is_default" boolean DEFAULT FALSE NOT NULL,
    PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX "uk_default_address" ON "addresses" ("user_id")
WHERE
    "is_default" = TRUE;

