metricstest -test.v -test.run=^TestIteration10A$ -agent-binary-path="cmd/agent/agent.exe" -binary-path="cmd/server/server.exe" -source-path=. -server-port=8080 -file-storage-path="\tmp\1.json" -database-dsn="postgres://postgres:mysecretpassword@localhost:5432/metrics?sslmode=disable"