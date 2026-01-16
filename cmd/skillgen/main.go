package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

	logger, _ := zap.NewProduction()
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
		cancel()
	}()

	githubClient := github.NewClient(cfg.GitHub.Token, logger)
	prFetcher := github.NewPRFetcher(githubClient, logger)
	llmClient := llm.NewClient(cfg.LLM.BaseURL, cfg.LLM.Model, cfg.LLM.Timeout, logger)
	llmAdapter := skillgen.NewLLMAdapter(llmClient)

	outputDir := cfg.SkillGen.OutputDir
	if outputDir == "" {
		outputDir = ".claude/skills"
	}

	generator := skillgen.NewGenerator(
		database.Get(),
		prFetcher,
		llmAdapter,
		logger,
		cfg.Repository.Owner,
		cfg.Repository.Name,
		outputDir,
	)

	cmd := os.Args[1]
	switch cmd {
	case "generate":
		if err := generator.Generate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating skills: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Skills generated successfully")

	case "sync":
		if err := generator.Sync(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error syncing PRs: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("PRs synced successfully")

	case "list":
		skills, err := generator.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing skills: %v\n", err)
			os.Exit(1)
		}
		if len(skills) == 0 {
			fmt.Println("No skills found")
		} else {
			fmt.Println("Generated skills:")
			for _, skill := range skills {
				fmt.Printf("  - %s\n", skill)
			}
		}

	case "update":
		if err := generator.Update(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating skills: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Skills updated successfully")

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`upgo - GitHub PR analysis to Claude Code Skills generator

Usage:
  upgo <command>

Commands:
  generate    Generate skill files from PR analysis
  sync        Sync PR data from GitHub
  list        List generated skills
  update      Update existing skills with new PR data
  help        Show this help message

Examples:
  upgo sync       # Fetch latest PRs from GitHub
  upgo generate   # Generate SKILL.md files from PR analysis
  upgo list       # Show all generated skills
  upgo update     # Update skills with new insights`)
}
