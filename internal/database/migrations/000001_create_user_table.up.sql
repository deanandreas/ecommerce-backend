CREATE TABLE "users" (
    "id" UUID DEFAULT gen_random_uuid(),
    "full_name" VARCHAR(255) NOT NULL,
    "birth" DATE NOT NULL,
    "phone" VARCHAR(255) NOT NULL,
    "email" VARCHAR(255) UNIQUE NOT NULL,
    "hash_password" VARCHAR(255) NOT NULL,
    "created_at" TIMESTAMP DEFAULT NOW(),
    "updated_at" TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY ("id")
);

CREATE INDEX "email_index" ON "users" ("email");
