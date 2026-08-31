CREATE TABLE "refresh_tokens"(
  "id" UUID DEFAULT gen_random_uuid(),
  "user_id" UUID NOT NULL,
  "hash_token" TEXT UNIQUE NOT NULL,
  "expiared_at" DATE NOT NULL,
  PRIMARY KEY("id"),
  CONSTRAINT fk_user FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE CASCADE
);
