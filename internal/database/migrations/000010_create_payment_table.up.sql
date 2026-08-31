CREATE TABLE "payments" (
    "id" uuid DEFAULT gen_random_uuid (),
    "user_id" uuid NOT NULL,
    "transaction_id" varchar(255) NOT NULL,
    "status" varchar(25) DEFAULT 'pending' CHECK ("status" IN ('pending', 'paid', 'cancelled')) NOT NULL,
    "order_id" uuid NOT NULL,
    "payment_gateway" varchar(255) NOT NULL,
    "amount_in_cent" int NOT NULL,
    "currency" varchar(255) NOT NULL,
    "payment_method" varchar(255) NOT NULL,
    "created_at" timestamp DEFAULT NOW(),
    PRIMARY KEY ("id")
);
