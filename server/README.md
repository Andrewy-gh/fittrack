## How to run in development

1. Copy `setenv.example.sh` to `setenv.sh` and modify it to your liking.

2. Run:
```bash
. setenv.sh
```

3. Run in the same terminal window:
```bash
docker compose up -d
```

4. Initialize the database:
> Note db container name is `db`
```bash
cat schema.sql | docker exec -i db psql -U ${DB_USER} -d ${DB_NAME}
```

5. Run in the same terminal window:
```bash
go run ./cmd/api
```