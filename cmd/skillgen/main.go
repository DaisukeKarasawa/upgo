package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"upgo/internal/analyzer"
	"upgo/internal/config"
	"upgo/internal/database"
	"upgo/internal/github"
	"upgo/internal/llm"
	"upgo/internal/skillgen"

	"go.uber.org/zap"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg, err := config.Load(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	if err := database.Connect(cfg.Database.Path, logger); err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := database.RunMigrations(logger); err != nil {
		fmt.Fprintf(os.Stderr, "Error running migrations: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nInterrupted. Shutting down...")
		cancel()
	}()

	githubClient := github.NewClient(cfg.GitHub.Token, logger)
	prFetcher := github.NewPRFetcher(githubClient, logger)
	llmClient := llm.NewClient(cfg.LLM.BaseURL, cfg.LLM.Model, cfg.LLM.Timeout, logger)

	prAnalyzer := analyzer.NewPRAnalyzer(database.Get(), llmClient, logger)
	generator := skillgen.NewGenerator(database.Get(), llmClient, logger, cfg.SkillGen.OutputDir)

	cmd := os.Args[1]
	switch cmd {
	case "sync":
		fmt.Printf("Syncing PRs from %s/%s (last 30 days)...\n", cfg.Repository.Owner, cfg.Repository.Name)
		since := time.Now().AddDate(0, 0, -30)
		count, err := syncPRs(ctx, cfg, prFetcher, logger, since)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error syncing PRs: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Synced %d PRs successfully\n", count)

	case "analyze":
		fmt.Println("Analyzing PRs...")
		count, err := prAnalyzer.AnalyzeRecentPRs(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error analyzing PRs: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Analyzed %d PRs successfully\n", count)

	case "generate":
		fmt.Println("Generating skills from PR analysis...")
		skills, err := generator.Generate(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating skills: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated %d skill files:\n", len(skills))
		for _, skill := range skills {
			fmt.Printf("  - %s\n", skill)
		}

	case "run":
		// Full pipeline: sync -> analyze -> generate
		fmt.Println("=== Running full pipeline ===")

		fmt.Printf("\n[1/3] Syncing PRs from %s/%s...\n", cfg.Repository.Owner, cfg.Repository.Name)
		since := time.Now().AddDate(0, 0, -30)
		syncCount, err := syncPRs(ctx, cfg, prFetcher, logger, since)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error syncing PRs: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Synced %d PRs\n", syncCount)

		fmt.Println("\n[2/3] Analyzing PRs...")
		analyzeCount, err := prAnalyzer.AnalyzeRecentPRs(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error analyzing PRs: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Analyzed %d PRs\n", analyzeCount)

		fmt.Println("\n[3/3] Generating skills...")
		skills, err := generator.Generate(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating skills: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated %d skill files\n", len(skills))

		fmt.Println("\n=== Pipeline complete ===")
		fmt.Printf("Skills are available in: %s\n", cfg.SkillGen.OutputDir)

	case "list":
		skills, err := generator.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing skills: %v\n", err)
			os.Exit(1)
		}
		if len(skills) == 0 {
			fmt.Println("No skills found. Run 'upgo run' to generate skills.")
		} else {
			fmt.Println("Generated skills:")
			for _, skill := range skills {
				fmt.Printf("  - %s\n", skill)
			}
		}

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func syncPRs(ctx context.Context, cfg *config.Config, prFetcher *github.PRFetcher, logger *zap.Logger, since time.Time) (int, error) {
	db := database.Get()

	// Get or create repository
	var repoID int
	err := db.QueryRow(
		"SELECT id FROM repositories WHERE owner = ? AND name = ?",
		cfg.Repository.Owner, cfg.Repository.Name,
	).Scan(&repoID)

	if err != nil {
		result, err := db.Exec(
			"INSERT INTO repositories (owner, name, last_synced_at) VALUES (?, ?, ?)",
			cfg.Repository.Owner, cfg.Repository.Name, nil,
		)
		if err != nil {
			return 0, fmt.Errorf("failed to create repository: %w", err)
		}
		id, _ := result.LastInsertId()
		repoID = int(id)
	}

	totalCount := 0

	for _, state := range []string{"open", "closed"} {
		prs, err := prFetcher.FetchPRsUpdatedSince(ctx, cfg.Repository.Owner, cfg.Repository.Name, state, since)
		if err != nil {
			logger.Warn("Failed to fetch PRs", zap.String("state", state), zap.Error(err))
			continue
		}

		for _, pr := range prs {
			prState := pr.GetState()
			if !pr.GetMergedAt().IsZero() {
				prState = "merged"
			}

			// Fetch comments for PR
			comments, err := prFetcher.FetchPRComments(ctx, cfg.Repository.Owner, cfg.Repository.Name, pr.GetNumber())
			if err != nil {
				logger.Warn("Failed to fetch comments", zap.Int("pr", pr.GetNumber()), zap.Error(err))
			}

			// Combine all comments
			var allComments string
			for _, c := range comments {
				allComments += fmt.Sprintf("**%s**: %s\n\n", c.GetUser().GetLogin(), c.GetBody())
			}

			// Fetch diff
			diff, err := prFetcher.FetchPRDiff(ctx, cfg.Repository.Owner, cfg.Repository.Name, pr.GetNumber())
			if err != nil {
				logger.Warn("Failed to fetch diff", zap.Int("pr", pr.GetNumber()), zap.Error(err))
			}

			// Save PR with comments and diff
			_, err = db.Exec(`
				INSERT INTO pull_requests
				(repository_id, github_id, title, body, state, author, created_at, updated_at, last_synced_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT(repository_id, github_id) DO UPDATE SET
				title = excluded.title,
				body = excluded.body,
				state = excluded.state,
				updated_at = excluded.updated_at,
				last_synced_at = excluded.last_synced_at`,
				repoID, pr.GetNumber(), pr.GetTitle(), pr.GetBody(), prState,
				pr.GetUser().GetLogin(), pr.GetCreatedAt().Time, pr.GetUpdatedAt().Time, time.Now(),
			)
			if err != nil {
				logger.Warn("Failed to save PR", zap.Int("number", pr.GetNumber()), zap.Error(err))
				continue
			}

			// Get PR ID
			var prID int
			db.QueryRow("SELECT id FROM pull_requests WHERE repository_id = ? AND github_id = ?", repoID, pr.GetNumber()).Scan(&prID)

			// Save comments
			if allComments != "" {
				db.Exec(`
					INSERT INTO pull_request_comments (pr_id, github_id, body, author, created_at, updated_at)
					VALUES (?, 0, ?, 'combined', ?, ?)
					ON CONFLICT(pr_id, github_id) DO UPDATE SET body = excluded.body`,
					prID, allComments, time.Now(), time.Now(),
				)
			}

			// Save diff
			if diff != "" {
				db.Exec(`
					INSERT INTO pull_request_diffs (pr_id, diff_text, file_path, created_at)
					VALUES (?, ?, 'all', ?)
					ON CONFLICT(pr_id, file_path) DO UPDATE SET diff_text = excluded.diff_text`,
					prID, diff, time.Now(),
				)
			}

			totalCount++
		}

		logger.Info("Synced PRs", zap.String("state", state), zap.Int("count", len(prs)))
	}

	// Update last sync time
	db.Exec("UPDATE repositories SET last_synced_at = ? WHERE id = ?", time.Now(), repoID)

	return totalCount, nil
}

func printUsage() {
	fmt.Println(`upgo - Go PR Analysis to Claude Code Skills generator

Usage:
  upgo <command>

Commands:
  sync        Sync PR data from GitHub (golang/go repository)
  analyze     Analyze PRs using LLM to extract Go insights
  generate    Generate SKILL.md files from analysis
  run         Run full pipeline (sync -> analyze -> generate)
  list        List generated skills
  help        Show this help message

Examples:
  upgo run        # Full pipeline: fetch PRs, analyze, generate skills
  upgo sync       # Only fetch latest PRs
  upgo analyze    # Only analyze fetched PRs
  upgo generate   # Only generate skills from existing analysis
  upgo list       # Show generated skills

Configuration:
  Edit config.yaml to set:
  - repository.owner/name: Target repository (default: golang/go)
  - skillgen.output_dir: Output directory for skills (default: skills/)

The generated skills can be used with Claude Code:
  cp -r skills/* ~/.claude/skills/`)
}
