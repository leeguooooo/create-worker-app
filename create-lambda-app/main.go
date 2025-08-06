package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/leeguooooo/create-lambda-app/internal/generator"
)

var (
	version = "1.0.0"
	bold    = color.New(color.Bold).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	cyan    = color.New(color.FgCyan).SprintFunc()
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "create-lambda-app [project-name]",
		Short: "Create a new Go Lambda function project",
		Long: bold("Create Lambda App") + " - A professional scaffolding tool for AWS Lambda functions in Go\n\n" +
			"Create production-ready serverless applications with best practices built-in:\n" +
			"  • Structured project layout with clean architecture\n" +
			"  • Multi-environment configuration (dev/staging/prod)\n" +
			"  • Comprehensive testing setup with mocks\n" +
			"  • CI/CD pipelines with GitHub Actions\n" +
			"  • OpenAPI documentation generation\n" +
			"  • Built-in middleware for logging, tracing, and error handling\n" +
			"  • SAM/CDK deployment configurations\n" +
			"  • Handler generators for common patterns",
		Version: version,
		RunE:    run,
	}

	rootCmd.Flags().StringP("name", "n", "", "Project name")
	rootCmd.Flags().StringP("description", "d", "", "Project description")
	rootCmd.Flags().BoolP("skip-git", "", false, "Skip git initialization")
	rootCmd.Flags().BoolP("skip-install", "", false, "Skip dependency installation")
	rootCmd.Flags().StringP("deployment", "", "", "Deployment tool (sam/cdk/serverless)")
	rootCmd.Flags().StringSliceP("features", "f", []string{}, "Features to include (api,dynamodb,sqs,sns,s3,cognito)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(red("Error:"), err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(bold("🚀 Create Lambda App"))
	fmt.Println(cyan("   Professional Go Lambda Function Generator"))
	fmt.Println()

	// Get project configuration
	config, err := getProjectConfig(cmd, args)
	if err != nil {
		return err
	}

	// Validate project path
	projectPath := filepath.Join(".", config.Name)
	if _, err := os.Stat(projectPath); err == nil {
		return fmt.Errorf("directory %s already exists", config.Name)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking directory %s: %w", config.Name, err)
	}

	// Create project
	fmt.Println(yellow("Creating project structure..."))
	if err := generator.Generate(config); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	// Initialize git
	if !config.SkipGit {
		fmt.Println(yellow("Initializing git repository..."))
		if err := generator.InitGit(projectPath); err != nil {
			fmt.Printf(red("Warning: ") + "Failed to initialize git: %v\n", err)
		}
	}

	// Install dependencies
	if !config.SkipInstall {
		fmt.Println(yellow("Installing dependencies..."))
		if err := generator.InstallDependencies(projectPath); err != nil {
			fmt.Printf(red("Warning: ") + "Failed to install dependencies: %v\n", err)
		}
	}

	// Success message
	fmt.Println()
	fmt.Println(green("✨ Successfully created project: ") + bold(config.Name))
	fmt.Println()
	fmt.Println(bold("Next steps:"))
	fmt.Printf("  cd %s\n", config.Name)
	if config.SkipInstall {
		fmt.Println("  go mod download")
	}
	fmt.Println("  make test")
	fmt.Println("  make run-local")
	fmt.Println()
	fmt.Println(bold("Available commands:"))
	fmt.Println("  make generate-handler  " + cyan("# Generate new Lambda handlers"))
	fmt.Println("  make build            " + cyan("# Build all Lambda functions"))
	fmt.Println("  make test             " + cyan("# Run tests with coverage"))
	fmt.Println("  make deploy-dev       " + cyan("# Deploy to development"))
	fmt.Println("  make deploy-prod      " + cyan("# Deploy to production"))
	fmt.Println()
	fmt.Println(bold("Documentation:"))
	fmt.Println("  • README.md           " + cyan("# Project overview and setup"))
	fmt.Println("  • docs/ARCHITECTURE.md " + cyan("# Architecture decisions"))
	fmt.Println("  • docs/DEPLOYMENT.md  " + cyan("# Deployment guide"))
	fmt.Println()

	return nil
}

func getProjectConfig(cmd *cobra.Command, args []string) (*generator.Config, error) {
	config := &generator.Config{
		Features: make(map[string]bool),
	}

	// Get project name
	if len(args) > 0 {
		config.Name = args[0]
	} else if name, _ := cmd.Flags().GetString("name"); name != "" {
		config.Name = name
	} else {
		if err := survey.AskOne(&survey.Input{
			Message: "Project name:",
			Help:    "The name of your Lambda project (e.g., my-api-service)",
		}, &config.Name, survey.WithValidator(survey.Required)); err != nil {
			return nil, err
		}
	}

	// Normalize and validate project name
	config.Name = strings.ToLower(strings.ReplaceAll(config.Name, " ", "-"))
	
	// Validate project name
	validName := regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	if !validName.MatchString(config.Name) {
		return nil, fmt.Errorf("project name must start with a letter and contain only lowercase letters, numbers, and hyphens")
	}

	// Get description
	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		config.Description = desc
	} else {
		if err := survey.AskOne(&survey.Input{
			Message: "Project description:",
			Default: fmt.Sprintf("AWS Lambda functions for %s", config.Name),
		}, &config.Description); err != nil {
			return nil, err
		}
	}

	// Get deployment tool
	if deployment, _ := cmd.Flags().GetString("deployment"); deployment != "" {
		config.DeploymentTool = deployment
	} else {
		if err := survey.AskOne(&survey.Select{
			Message: "Choose deployment tool:",
			Options: []string{
				"sam (AWS Serverless Application Model - AWS native, simple configuration)",
				"cdk (AWS Cloud Development Kit - TypeScript/Python, programmable infrastructure)",
				"serverless (Serverless Framework - Multi-cloud, large plugin ecosystem)",
				"terraform (HashiCorp Terraform - Multi-provider, declarative infrastructure)",
			},
			Default: "sam (AWS Serverless Application Model - AWS native, simple configuration)",
		}, &config.DeploymentTool); err != nil {
			return nil, err
		}
		// Extract the short form
		config.DeploymentTool = strings.Split(config.DeploymentTool, " ")[0]
	}

	// Get features
	if features, _ := cmd.Flags().GetStringSlice("features"); len(features) > 0 {
		for _, f := range features {
			config.Features[f] = true
		}
	} else {
		selectedFeatures := []string{}
		if err := survey.AskOne(&survey.MultiSelect{
			Message: "Select features to include:",
			Options: []string{
				"api (API Gateway - REST APIs with routing and validation)",
				"dynamodb (DynamoDB - NoSQL database for user/room data)",
				"sqs (SQS - Message queue for async processing)",
				"sns (SNS - Pub/sub messaging for notifications)",
				"s3 (S3 - Object storage for files/media)",
				"cognito (Cognito - User authentication and authorization)",
				"secrets (Secrets Manager - Store API keys and credentials)",
				"eventbridge (EventBridge - Event-driven triggers)",
				"stepfunctions (Step Functions - Workflow orchestration)",
			},
			Default: []string{"api (API Gateway - REST APIs with routing and validation)"},
		}, &selectedFeatures); err != nil {
			return nil, err
		}

		for _, f := range selectedFeatures {
			feature := strings.Split(f, " ")[0]
			config.Features[feature] = true
		}
	}

	// Additional options
	config.SkipGit, _ = cmd.Flags().GetBool("skip-git")
	config.SkipInstall, _ = cmd.Flags().GetBool("skip-install")

	// Architecture preferences
	if err := survey.AskOne(&survey.Select{
		Message: "Choose project structure:",
		Options: []string{
			"clean (Clean Architecture with use cases)",
			"simple (Simple handler-based structure)",
			"ddd (Domain-Driven Design)",
		},
		Default: "clean (Clean Architecture with use cases)",
	}, &config.Architecture); err != nil {
		return nil, err
	}
	// Extract the short form
	config.Architecture = strings.Split(config.Architecture, " ")[0]

	// Testing framework
	if err := survey.AskOne(&survey.Select{
		Message: "Choose testing approach:",
		Options: []string{
			"testify (Assertions and mocks)",
			"standard (Standard library only)",
			"ginkgo (BDD-style testing)",
		},
		Default: "testify (Assertions and mocks)",
	}, &config.TestingFramework); err != nil {
		return nil, err
	}
	// Extract the short form
	config.TestingFramework = strings.Split(config.TestingFramework, " ")[0]

	return config, nil
}