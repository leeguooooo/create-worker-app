package templates

// CI/CD templates

const GitHubActionsCI = `name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  GO_VERSION: '1.21'
  AWS_REGION: us-east-1

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          
      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-
            
      - name: Install dependencies
        run: go mod download
        
      - name: Run tests
        run: make test
        
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
          flags: unittests
          name: codecov-umbrella
          
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          
  security:
    name: Security Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Run Gosec Security Scanner
        uses: securego/gosec@master
        with:
          args: ./...
          
  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [test, lint]
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          
      - name: Build Lambda functions
        run: make build
        
      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: lambda-functions
          path: build/
          retention-days: 7
          
  {{- if .HasFeature "api" }}
  api-docs:
    name: Generate API Documentation
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          
      - name: Install swag
        run: go install github.com/swaggo/swag/cmd/swag@latest
        
      - name: Generate Swagger docs
        run: swag init -g ./internal/interfaces/api/router.go -o ./docs
        
      - name: Upload API docs
        uses: actions/upload-artifact@v3
        with:
          name: api-docs
          path: docs/
  {{- end }}
`

const GitHubActionsDeploy = `name: Deploy

on:
  push:
    branches:
      - main
      - develop
  workflow_dispatch:
    inputs:
      environment:
        description: 'Environment to deploy to'
        required: true
        type: choice
        options:
          - dev
          - staging
          - prod

env:
  GO_VERSION: '1.21'
  AWS_REGION: us-east-1

jobs:
  build:
    name: Build
    runs-on: ubuntu-latest
    outputs:
      version: ${{ steps.version.outputs.version }}
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          
      - name: Generate version
        id: version
        run: |
          VERSION=$(date +%Y%m%d%H%M%S)-${GITHUB_SHA::7}
          echo "version=$VERSION" >> $GITHUB_OUTPUT
          echo "Version: $VERSION"
          
      - name: Build Lambda functions
        run: make build
        
      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: lambda-functions-${{ steps.version.outputs.version }}
          path: build/
          retention-days: 30

  deploy-dev:
    name: Deploy to Development
    runs-on: ubuntu-latest
    needs: build
    if: github.ref == 'refs/heads/develop' || (github.event_name == 'workflow_dispatch' && github.event.inputs.environment == 'dev')
    environment:
      name: development
      url: ${{ steps.deploy.outputs.api_url }}
    steps:
      - uses: actions/checkout@v4
      
      - name: Download artifacts
        uses: actions/download-artifact@v3
        with:
          name: lambda-functions-${{ needs.build.outputs.version }}
          path: build/
          
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ env.AWS_REGION }}
          
      {{- if eq .DeploymentTool "sam" }}
      - name: Install SAM CLI
        uses: aws-actions/setup-sam@v2
        
      - name: Deploy with SAM
        id: deploy
        run: |
          sam deploy \
            --config-env dev \
            --parameter-overrides \
              Version=${{ needs.build.outputs.version }} \
            --no-confirm-changeset \
            --no-fail-on-empty-changeset
          
          # Get stack outputs
          API_URL=$(aws cloudformation describe-stacks \
            --stack-name {{.Name}}-dev \
            --query "Stacks[0].Outputs[?OutputKey=='ApiUrl'].OutputValue" \
            --output text)
          echo "api_url=$API_URL" >> $GITHUB_OUTPUT
      {{- else if eq .DeploymentTool "cdk" }}
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          
      - name: Install CDK dependencies
        working-directory: ./cdk
        run: npm ci
        
      - name: Deploy with CDK
        id: deploy
        working-directory: ./cdk
        run: |
          npm run deploy:dev -- --require-approval never
      {{- else if eq .DeploymentTool "serverless" }}
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          
      - name: Install Serverless Framework
        run: npm install -g serverless
        
      - name: Deploy with Serverless
        id: deploy
        run: |
          serverless deploy --stage dev
      {{- else if eq .DeploymentTool "terraform" }}
      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v2
        with:
          terraform_version: 1.5.0
          
      - name: Terraform Init
        working-directory: ./terraform
        run: terraform init
        
      - name: Terraform Apply
        id: deploy
        working-directory: ./terraform
        run: |
          terraform apply -auto-approve -var-file=environments/dev.tfvars
      {{- end }}
      
      - name: Tag deployment
        run: |
          git tag -a "dev-${{ needs.build.outputs.version }}" -m "Deploy to dev: ${{ needs.build.outputs.version }}"
          git push origin "dev-${{ needs.build.outputs.version }}"

  deploy-staging:
    name: Deploy to Staging
    runs-on: ubuntu-latest
    needs: [build, deploy-dev]
    if: github.ref == 'refs/heads/main' || (github.event_name == 'workflow_dispatch' && github.event.inputs.environment == 'staging')
    environment:
      name: staging
      url: ${{ steps.deploy.outputs.api_url }}
    steps:
      - uses: actions/checkout@v4
      
      - name: Download artifacts
        uses: actions/download-artifact@v3
        with:
          name: lambda-functions-${{ needs.build.outputs.version }}
          path: build/
          
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ env.AWS_REGION }}
          
      {{- if eq .DeploymentTool "sam" }}
      - name: Install SAM CLI
        uses: aws-actions/setup-sam@v2
        
      - name: Deploy with SAM
        id: deploy
        run: |
          sam deploy \
            --config-env staging \
            --parameter-overrides \
              Version=${{ needs.build.outputs.version }} \
            --no-confirm-changeset \
            --no-fail-on-empty-changeset
      {{- else if eq .DeploymentTool "cdk" }}
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          
      - name: Install CDK dependencies
        working-directory: ./cdk
        run: npm ci
        
      - name: Deploy with CDK
        id: deploy
        working-directory: ./cdk
        run: |
          npm run deploy:staging -- --require-approval never
      {{- else if eq .DeploymentTool "serverless" }}
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          
      - name: Install Serverless Framework
        run: npm install -g serverless
        
      - name: Deploy with Serverless
        id: deploy
        run: |
          serverless deploy --stage staging
      {{- else if eq .DeploymentTool "terraform" }}
      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v2
        with:
          terraform_version: 1.5.0
          
      - name: Terraform Init
        working-directory: ./terraform
        run: terraform init
        
      - name: Terraform Apply
        id: deploy
        working-directory: ./terraform
        run: |
          terraform apply -auto-approve -var-file=environments/staging.tfvars
      {{- end }}
      
      - name: Run integration tests
        run: |
          # Add your integration test commands here
          echo "Running integration tests..."
          
      - name: Tag deployment
        run: |
          git tag -a "staging-${{ needs.build.outputs.version }}" -m "Deploy to staging: ${{ needs.build.outputs.version }}"
          git push origin "staging-${{ needs.build.outputs.version }}"

  deploy-prod:
    name: Deploy to Production
    runs-on: ubuntu-latest
    needs: [build, deploy-staging]
    if: github.event_name == 'workflow_dispatch' && github.event.inputs.environment == 'prod'
    environment:
      name: production
      url: ${{ steps.deploy.outputs.api_url }}
    steps:
      - uses: actions/checkout@v4
      
      - name: Download artifacts
        uses: actions/download-artifact@v3
        with:
          name: lambda-functions-${{ needs.build.outputs.version }}
          path: build/
          
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.PROD_AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.PROD_AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ env.AWS_REGION }}
          
      {{- if eq .DeploymentTool "sam" }}
      - name: Install SAM CLI
        uses: aws-actions/setup-sam@v2
        
      - name: Deploy with SAM
        id: deploy
        run: |
          sam deploy \
            --config-env prod \
            --parameter-overrides \
              Version=${{ needs.build.outputs.version }} \
            --no-fail-on-empty-changeset
      {{- else if eq .DeploymentTool "cdk" }}
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          
      - name: Install CDK dependencies
        working-directory: ./cdk
        run: npm ci
        
      - name: Deploy with CDK
        id: deploy
        working-directory: ./cdk
        run: |
          npm run deploy:prod
      {{- else if eq .DeploymentTool "serverless" }}
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          
      - name: Install Serverless Framework
        run: npm install -g serverless
        
      - name: Deploy with Serverless
        id: deploy
        run: |
          serverless deploy --stage prod
      {{- else if eq .DeploymentTool "terraform" }}
      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v2
        with:
          terraform_version: 1.5.0
          
      - name: Terraform Init
        working-directory: ./terraform
        run: terraform init
        
      - name: Terraform Plan
        working-directory: ./terraform
        run: |
          terraform plan -var-file=environments/prod.tfvars -out=tfplan
          
      - name: Terraform Apply
        id: deploy
        working-directory: ./terraform
        run: |
          terraform apply tfplan
      {{- end }}
      
      - name: Create release
        uses: actions/create-release@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          tag_name: v${{ needs.build.outputs.version }}
          release_name: Release ${{ needs.build.outputs.version }}
          body: |
            Production deployment of version ${{ needs.build.outputs.version }}
            
            ## Changes
            - Deployed to production environment
            - All tests passed
            
            ## Deployment Info
            - Environment: Production
            - Region: ${{ env.AWS_REGION }}
            - Timestamp: ${{ github.event.head_commit.timestamp }}
          draft: false
          prerelease: false
`