metricstest -test.v -test.run=^TestIteration10$ -agent-binary-path="cmd/agent/agent.exe" -binary-path="cmd/server/server.exe" -source-path=. -server-port=8080 -file-storage-path="\tmp\1.json" -database-dsn="host=localhost port=5432 user=postgres password=mysecretpassword dbname=metrics sslmode=disable"TestIteration9