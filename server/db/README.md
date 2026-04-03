### Database Migrations

`server/db` の SQLite スキーマは `sqlc` と `goose` で管理します。`sqlc` の schema 入力は `server/db/migrations/` を直接参照しており、アプリ起動時にも未適用 migration を自動で適用します。

```bash
# 新しい migration を作成
go run ./cmd/dbmigrate create add_some_column

# 状態確認
go run ./cmd/dbmigrate status

# 手動適用
go run ./cmd/dbmigrate up
```

新規 migration は `00001_xxx.sql` のようなゼロ埋め連番で作成されます。これは `sqlc` が migration ディレクトリを順番に読む都合に合わせています。