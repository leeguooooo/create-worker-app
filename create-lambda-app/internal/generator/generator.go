package generator

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/leeguooooo/create-lambda-app/internal/templates"
)

// getGitHubUsername attempts to get the GitHub username from git config
func getGitHubUsername() string {
	cmd := exec.Command("git", "config", "--get", "user.name")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		username := strings.TrimSpace(string(output))
		// Convert to lowercase and replace spaces with hyphens
		username = strings.ToLower(strings.ReplaceAll(username, " ", "-"))
		return username
	}
	return "myusername"
}

// Generate creates a new Lambda project based on the configuration
func Generate(config *Config) error {
	// Set default module name if not provided
	if config.Module == "" {
		config.Module = fmt.Sprintf("github.com/%s/%s", getGitHubUsername(), config.Name)
	}

	// Create project directory
	projectPath := filepath.Join(".", config.Name)
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Generate base structure based on architecture
	switch config.Architecture {
	case "clean":
		if err := generateCleanArchitecture(projectPath, config); err != nil {
			return err
		}
	case "simple":
		if err := generateSimpleStructure(projectPath, config); err != nil {
			return err
		}
	case "ddd":
		if err := generateDDDStructure(projectPath, config); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown architecture: %s", config.Architecture)
	}

	// Generate common files
	if err := generateCommonFiles(projectPath, config); err != nil {
		return err
	}

	// Generate deployment configuration
	if err := generateDeploymentConfig(projectPath, config); err != nil {
		return err
	}

	// Generate feature-specific code
	if err := generateFeatures(projectPath, config); err != nil {
		return err
	}

	// Create metadata file
	if err := createMetadataFile(projectPath, config); err != nil {
		return err
	}

	return nil
}

func generateCleanArchitecture(projectPath string, config *Config) error {
	dirs := []string{
		"cmd",
		"internal/domain/entities",
		"internal/domain/repositories",
		"internal/domain/services",
		"internal/usecases",
		"internal/interfaces/lambda",
		"internal/interfaces/api",
		"internal/infrastructure/database",
		"internal/infrastructure/aws",
		"internal/infrastructure/config",
		"pkg/logger",
		"pkg/errors",
		"pkg/middleware",
		"test/unit",
		"test/integration",
		"test/e2e",
		"test/mocks",
		"docs",
		"scripts",
		"deployments",
	}

	// Create directories
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(projectPath, dir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate Clean Architecture specific files
	files := map[string]string{
		"internal/domain/entities/base.go":          templates.CleanBaseEntity,
		"internal/domain/repositories/interfaces.go": templates.CleanRepositoryInterface,
		"internal/usecases/interfaces.go":           templates.CleanUseCaseInterface,
		"internal/interfaces/lambda/handler.go":     templates.CleanLambdaHandler,
		"internal/infrastructure/config/config.go":  templates.CleanConfig,
		"pkg/logger/logger.go":                      templates.Logger,
		"pkg/errors/errors.go":                      templates.CustomErrors,
		"pkg/middleware/middleware.go":              templates.Middleware,
	}

	for path, content := range files {
		if err := generateFile(filepath.Join(projectPath, path), content, config); err != nil {
			return err
		}
	}

	return nil
}

func generateSimpleStructure(projectPath string, config *Config) error {
	dirs := []string{
		"handlers",
		"models",
		"services",
		"utils",
		"config",
		"test",
		"scripts",
		"deployments",
	}

	// Create directories
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(projectPath, dir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate simple structure specific files
	files := map[string]string{
		"handlers/main.go":     templates.SimpleHandler,
		"models/models.go":     templates.SimpleModels,
		"services/service.go":  templates.SimpleService,
		"utils/utils.go":       templates.SimpleUtils,
		"config/config.go":     templates.SimpleConfig,
	}

	for path, content := range files {
		if err := generateFile(filepath.Join(projectPath, path), content, config); err != nil {
			return err
		}
	}

	return nil
}

func generateDDDStructure(projectPath string, config *Config) error {
	dirs := []string{
		"cmd",
		"domain/aggregate",
		"domain/entity",
		"domain/valueobject",
		"domain/repository",
		"domain/service",
		"domain/event",
		"application/command",
		"application/query",
		"application/handler",
		"infrastructure/persistence",
		"infrastructure/messaging",
		"infrastructure/config",
		"interfaces/lambda",
		"interfaces/api",
		"test",
		"scripts",
		"deployments",
	}

	// Create directories
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(projectPath, dir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate DDD specific files
	files := map[string]string{
		"domain/aggregate/base.go":              templates.DDDAggregateBase,
		"domain/entity/base.go":                 templates.DDDEntityBase,
		"domain/valueobject/base.go":            templates.DDDValueObject,
		"domain/repository/interfaces.go":       templates.DDDRepository,
		"domain/event/base.go":                  templates.DDDEvent,
		"application/command/base.go":           templates.DDDCommand,
		"application/query/base.go":             templates.DDDQuery,
		"infrastructure/persistence/dynamodb.go": templates.DDDPersistence,
	}

	for path, content := range files {
		if err := generateFile(filepath.Join(projectPath, path), content, config); err != nil {
			return err
		}
	}

	return nil
}

func generateCommonFiles(projectPath string, config *Config) error {
	files := map[string]string{
		"go.mod":              templates.GoMod,
		"Makefile":            templates.Makefile,
		"README.md":           templates.README,
		".gitignore":          templates.GitIgnore,
		".env.example":        templates.EnvExample,
		"docker-compose.yml":  templates.DockerCompose,
		"Dockerfile":          templates.Dockerfile,
		".github/workflows/ci.yml": templates.GitHubActionsCI,
		".github/workflows/deploy.yml": templates.GitHubActionsDeploy,
		"docs/ARCHITECTURE.md": templates.ArchitectureDoc,
		"docs/DEPLOYMENT.md":   templates.DeploymentDoc,
		"docs/API.md":         templates.APIDoc,
		"scripts/generate-handler.go": templates.HandlerGenerator,
		"scripts/local-setup.sh": templates.LocalSetupScript,
		"test/testutils/utils.go": templates.TestUtils,
	}

	// Create .github/workflows directory
	if err := os.MkdirAll(filepath.Join(projectPath, ".github/workflows"), 0755); err != nil {
		return err
	}

	for path, content := range files {
		if err := generateFile(filepath.Join(projectPath, path), content, config); err != nil {
			return err
		}
	}

	// Make scripts executable
	scripts := []string{"scripts/local-setup.sh"}
	for _, script := range scripts {
		if err := os.Chmod(filepath.Join(projectPath, script), 0755); err != nil {
			return err
		}
	}

	return nil
}

func generateDeploymentConfig(projectPath string, config *Config) error {
	switch config.DeploymentTool {
	case "sam":
		return generateSAMConfig(projectPath, config)
	case "cdk":
		return generateCDKConfig(projectPath, config)
	case "serverless":
		return generateServerlessConfig(projectPath, config)
	case "terraform":
		return generateTerraformConfig(projectPath, config)
	default:
		return fmt.Errorf("unknown deployment tool: %s", config.DeploymentTool)
	}
}

func generateSAMConfig(projectPath string, config *Config) error {
	files := map[string]string{
		"template.yaml":            templates.SAMTemplate,
		"samconfig.toml":          templates.SAMConfig,
		"buildspec.yml":           templates.BuildSpec,
		"deployments/dev.yaml":    templates.SAMEnvDev,
		"deployments/staging.yaml": templates.SAMEnvStaging,
		"deployments/prod.yaml":   templates.SAMEnvProd,
	}

	for path, content := range files {
		if err := generateFile(filepath.Join(projectPath, path), content, config); err != nil {
			return err
		}
	}

	return nil
}

func generateCDKConfig(projectPath string, config *Config) error {
	// Create CDK app structure
	cdkDir := filepath.Join(projectPath, "cdk")
	if err := os.MkdirAll(filepath.Join(cdkDir, "lib"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(cdkDir, "bin"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(cdkDir, "test"), 0755); err != nil {
		return err
	}

	files := map[string]string{
		"cdk/cdk.json":           templates.CDKConfig,
		"cdk/tsconfig.json":      templates.CDKTSConfig,
		"cdk/package.json":       templates.CDKPackageJSON,
		"cdk/lib/stack.ts":       templates.CDKStack,
		"cdk/bin/app.ts":         templates.CDKApp,
		"cdk/test/stack.test.ts": templates.CDKTest,
	}

	for path, content := range files {
		if err := generateFile(filepath.Join(projectPath, path), content, config); err != nil {
			return err
		}
	}

	return nil
}

func generateServerlessConfig(projectPath string, config *Config) error {
	files := map[string]string{
		"serverless.yml":             templates.ServerlessYML,
		"serverless.env.yml":         templates.ServerlessEnv,
		"deployments/dev.yml":        templates.ServerlessEnvDev,
		"deployments/staging.yml":    templates.ServerlessEnvStaging,
		"deployments/production.yml": templates.ServerlessEnvProd,
	}

	for path, content := range files {
		if err := generateFile(filepath.Join(projectPath, path), content, config); err != nil {
			return err
		}
	}

	return nil
}

func generateTerraformConfig(projectPath string, config *Config) error {
	// Create Terraform structure
	tfDir := filepath.Join(projectPath, "terraform")
	if err := os.MkdirAll(filepath.Join(tfDir, "modules/lambda"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(tfDir, "environments"), 0755); err != nil {
		return err
	}

	files := map[string]string{
		"terraform/main.tf":                  templates.TerraformMain,
		"terraform/variables.tf":             templates.TerraformVariables,
		"terraform/outputs.tf":               templates.TerraformOutputs,
		"terraform/versions.tf":              templates.TerraformVersions,
		"terraform/environments/dev.tfvars":  templates.TerraformEnvDev,
		"terraform/environments/prod.tfvars": templates.TerraformEnvProd,
		"terraform/modules/lambda/main.tf":   templates.TerraformLambdaModule,
	}

	for path, content := range files {
		if err := generateFile(filepath.Join(projectPath, path), content, config); err != nil {
			return err
		}
	}

	return nil
}

func generateFeatures(projectPath string, config *Config) error {
	// API Gateway
	if config.HasFeature("api") {
		if err := generateAPIFeature(projectPath, config); err != nil {
			return err
		}
	}

	// DynamoDB
	if config.HasFeature("dynamodb") {
		if err := generateDynamoDBFeature(projectPath, config); err != nil {
			return err
		}
	}

	// SQS
	if config.HasFeature("sqs") {
		if err := generateSQSFeature(projectPath, config); err != nil {
			return err
		}
	}

	// Additional features...
	// SNS, S3, Cognito, EventBridge, etc.

	return nil
}

func generateAPIFeature(projectPath string, config *Config) error {
	var files map[string]string

	switch config.Architecture {
	case "clean":
		files = map[string]string{
			"internal/interfaces/api/router.go":     templates.CleanAPIRouter,
			"internal/interfaces/api/handlers.go":   templates.CleanAPIHandlers,
			"internal/interfaces/api/middleware.go": templates.CleanAPIMiddleware,
			"internal/interfaces/api/responses.go":  templates.CleanAPIResponses,
		}
	case "simple":
		files = map[string]string{
			"handlers/api.go":       templates.SimpleAPIHandler,
			"models/api_models.go":  templates.SimpleAPIModels,
			"utils/api_utils.go":    templates.SimpleAPIUtils,
		}
	case "ddd":
		files = map[string]string{
			"interfaces/api/router.go":            templates.DDDAPIRouter,
			"interfaces/api/handlers.go":          templates.DDDAPIHandlers,
			"application/handler/api_handler.go":  templates.DDDAPIApplicationHandler,
		}
	}

	for path, content := range files {
		if err := generateFile(filepath.Join(projectPath, path), content, config); err != nil {
			return err
		}
	}

	// Generate OpenAPI spec
	if err := generateFile(filepath.Join(projectPath, "docs/openapi.yaml"), templates.OpenAPISpec, config); err != nil {
		return err
	}

	return nil
}

func generateDynamoDBFeature(projectPath string, config *Config) error {
	var files map[string]string

	switch config.Architecture {
	case "clean":
		files = map[string]string{
			"internal/infrastructure/database/dynamodb.go":     templates.CleanDynamoDBClient,
			"internal/infrastructure/database/repository.go":   templates.CleanDynamoDBRepository,
			"internal/domain/repositories/user_repository.go":  templates.CleanUserRepository,
		}
	case "simple":
		files = map[string]string{
			"services/dynamodb.go":    templates.SimpleDynamoDBService,
			"models/dynamo_models.go": templates.SimpleDynamoDBModels,
		}
	case "ddd":
		files = map[string]string{
			"infrastructure/persistence/dynamodb_repository.go": templates.DDDDynamoDBRepository,
			"domain/repository/user_repository.go":              templates.DDDUserRepository,
		}
	}

	for path, content := range files {
		if err := generateFile(filepath.Join(projectPath, path), content, config); err != nil {
			return err
		}
	}

	return nil
}

func generateSQSFeature(projectPath string, config *Config) error {
	var files map[string]string

	switch config.Architecture {
	case "clean":
		files = map[string]string{
			"internal/infrastructure/aws/sqs.go":      templates.CleanSQSClient,
			"internal/interfaces/lambda/sqs_handler.go": templates.CleanSQSHandler,
			"internal/usecases/process_message.go":     templates.CleanProcessMessageUseCase,
		}
	case "simple":
		files = map[string]string{
			"handlers/sqs.go":      templates.SimpleSQSHandler,
			"services/sqs.go":      templates.SimpleSQSService,
			"models/sqs_models.go": templates.SimpleSQSModels,
		}
	case "ddd":
		files = map[string]string{
			"interfaces/lambda/sqs_handler.go":        templates.DDDSQSHandler,
			"infrastructure/messaging/sqs_client.go":  templates.DDDSQSClient,
			"application/handler/message_handler.go":  templates.DDDMessageHandler,
		}
	}

	for path, content := range files {
		if err := generateFile(filepath.Join(projectPath, path), content, config); err != nil {
			return err
		}
	}

	return nil
}

func generateFile(path, templateContent string, config *Config) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Special handling for GitHub Actions files - don't parse as templates
	if strings.Contains(path, ".github/workflows/") {
		// Just write the content directly without template parsing
		if err := os.WriteFile(path, []byte(templateContent), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
		return nil
	}

	// Parse and execute template
	tmpl, err := template.New(filepath.Base(path)).Parse(templateContent)
	if err != nil {
		return fmt.Errorf("failed to parse template for %s: %w", path, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return fmt.Errorf("failed to execute template for %s: %w", path, err)
	}

	// Write file
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

func createMetadataFile(projectPath string, config *Config) error {
	metadata := fmt.Sprintf(`{
  "generator": "create-lambda-app",
  "version": "1.0.0",
  "created": "%s",
  "architecture": "%s",
  "deployment": "%s",
  "features": %v,
  "testing": "%s"
}`, time.Now().Format(time.RFC3339), config.Architecture, config.DeploymentTool, config.GetEnabledFeatures(), config.TestingFramework)

	return os.WriteFile(filepath.Join(projectPath, ".create-lambda-app"), []byte(metadata), 0644)
}

// InitGit initializes a git repository in the project directory
func InitGit(projectPath string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Add all files
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add files to git: %w", err)
	}

	// Initial commit
	cmd = exec.Command("git", "commit", "-m", "Initial commit from create-lambda-app")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}
	return nil
}

// InstallDependencies runs go mod download in the project directory
func InstallDependencies(projectPath string) error {
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}