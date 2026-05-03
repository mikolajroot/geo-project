-- reverse: create "sessions" table
DROP TABLE "sessions";
-- reverse: drop index "user_login" from table: "users"
CREATE INDEX "user_login" ON "users" ("login");
