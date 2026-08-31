CREATE TABLE "categories" (
    "id" UUID DEFAULT gen_random_uuid(),
    "name" VARCHAR(255) NOT NULL,
    "slug" VARCHAR(255) UNIQUE NOT NULL,
    "created_at" TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY ("id")
);
