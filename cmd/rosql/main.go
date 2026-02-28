package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gennadis/rosql/internal/db"
	"github.com/gennadis/rosql/internal/output"
	"github.com/gennadis/rosql/internal/query"
	"github.com/joho/godotenv"
)

const (
	defaultLimit = 50
	queryTimeout = 5 * time.Second
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("failed to read .env file: %v", err)
		os.Exit(1)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Println("DATABASE_URL is required")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.New(ctx, dsn)
	if err != nil {
		fmt.Printf("database connection error: %v", err)
		os.Exit(1)
	}
	defer database.Close()

	fmt.Println("connected in read only mode")
	runRepl(ctx, database)
}

func runRepl(ctx context.Context, database *db.DB) {
	r := bufio.NewReader(os.Stdin)
	var builder strings.Builder

	for {
		if builder.Len() == 0 {
			fmt.Print("ROsql> ")
		} else {
			fmt.Print("  ...> ")
		}

		line, err := r.ReadString('\n')
		if err != nil {
			fmt.Printf("read error: %v", err)
			return
		}

		line = strings.TrimSpace(line)

		// handle exit gracefully
		if line == "exit" || line == "quit" {
			return
		}

		if builder.Len() > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(line)

		// execute query only when user finishes statement with ';'
		if !strings.HasSuffix(line, ";") {
			continue
		}

		queryStr := strings.TrimRight(builder.String(), ";")

		builder.Reset()

		if strings.TrimSpace(queryStr) == "" {
			continue
		}

		if err := query.ValidateReadOnly(queryStr); err != nil {
			fmt.Printf("validating read only query: %v", err)
			continue
		}

		ctxQuery, cancel := context.WithTimeout(ctx, queryTimeout)
		start := time.Now()

		result, err := database.Query(ctxQuery, queryStr, defaultLimit)
		cancel()

		duration := time.Since(start)

		if err != nil {
			fmt.Printf("querying database: query: %q, err: %v\n", queryStr, err)
			continue
		}

		output.PrintTable(result.Columns, result.Rows)

		fmt.Printf("%d rows in %v\n", len(result.Rows), duration)
	}
}
