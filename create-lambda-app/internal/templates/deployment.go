package templates

// Deployment configuration templates

const SAMTemplate = `AWSTemplateFormatVersion: '2010-09-09'
Transform: AWS::Serverless-2016-10-31
Description: >
  {{.Name}}
  
  {{.Description}}

# Global values that are applied to all resources
Globals:
  Function:
    Timeout: 30
    MemorySize: 512
    Runtime: provided.al2023
    Architectures:
      - x86_64
    Environment:
      Variables:
        APP_NAME: {{.Name}}
        APP_ENV: !Ref Environment
        LOG_LEVEL: !Ref LogLevel
        AWS_XRAY_TRACING_NAME: {{.Name}}
        _X_AMZN_TRACE_ID: !Ref AWS::NoValue
    Tracing: Active
    Tags:
      Application: {{.Name}}
      Environment: !Ref Environment

Parameters:
  Environment:
    Type: String
    Default: dev
    AllowedValues:
      - dev
      - staging
      - prod
    Description: Deployment environment

  LogLevel:
    Type: String
    Default: info
    AllowedValues:
      - debug
      - info
      - warn
      - error
    Description: Application log level

Resources:
  {{- if .HasFeature "api" }}
  # API Gateway
  ApiGateway:
    Type: AWS::Serverless::Api
    Properties:
      StageName: !Ref Environment
      TracingEnabled: true
      Cors:
        AllowMethods: "'*'"
        AllowHeaders: "'*'"
        AllowOrigin: "'*'"
      Auth:
        DefaultAuthorizer: NONE
        {{- if .HasFeature "cognito" }}
        Authorizers:
          CognitoAuthorizer:
            UserPoolArn: !GetAtt CognitoUserPool.Arn
        {{- end }}
      DefinitionBody:
        Fn::Transform:
          Name: AWS::Include
          Parameters:
            Location: ./docs/openapi.yaml
  {{- end }}

  # Lambda Functions
  {{- if eq .Architecture "clean" }}
  UserFunction:
    Type: AWS::Serverless::Function
    Properties:
      FunctionName: !Sub ${AWS::StackName}-user-handler
      CodeUri: build/
      Handler: user/bootstrap
      {{- if .HasFeature "api" }}
      Events:
        CreateUser:
          Type: Api
          Properties:
            RestApiId: !Ref ApiGateway
            Path: /users
            Method: POST
        GetUser:
          Type: Api
          Properties:
            RestApiId: !Ref ApiGateway
            Path: /users/{id}
            Method: GET
        ListUsers:
          Type: Api
          Properties:
            RestApiId: !Ref ApiGateway
            Path: /users
            Method: GET
        UpdateUser:
          Type: Api
          Properties:
            RestApiId: !Ref ApiGateway
            Path: /users/{id}
            Method: PUT
        DeleteUser:
          Type: Api
          Properties:
            RestApiId: !Ref ApiGateway
            Path: /users/{id}
            Method: DELETE
      {{- end }}
      Environment:
        Variables:
          {{- if .HasFeature "dynamodb" }}
          DYNAMODB_TABLE_NAME: !Ref UserTable
          {{- end }}
      Policies:
        - AWSLambdaBasicExecutionRole
        {{- if .HasFeature "dynamodb" }}
        - DynamoDBCrudPolicy:
            TableName: !Ref UserTable
        {{- end }}
  {{- end }}

  {{- if .HasFeature "sqs" }}
  MessageProcessorFunction:
    Type: AWS::Serverless::Function
    Properties:
      FunctionName: !Sub ${AWS::StackName}-message-processor
      CodeUri: build/
      Handler: message-processor/bootstrap
      Events:
        MySQSEvent:
          Type: SQS
          Properties:
            Queue: !GetAtt MessageQueue.Arn
            BatchSize: 10
      Environment:
        Variables:
          SQS_QUEUE_URL: !Ref MessageQueue
      Policies:
        - AWSLambdaBasicExecutionRole
        - SQSPollerPolicy:
            QueueName: !GetAtt MessageQueue.QueueName
  {{- end }}

  {{- if .HasFeature "eventbridge" }}
  EventHandlerFunction:
    Type: AWS::Serverless::Function
    Properties:
      FunctionName: !Sub ${AWS::StackName}-event-handler
      CodeUri: build/
      Handler: event-handler/bootstrap
      Events:
        EventBridgeRule:
          Type: EventBridgeRule
          Properties:
            Pattern:
              source:
                - custom.{{.Name}}
              detail-type:
                - UserCreated
                - UserUpdated
                - UserDeleted
      Policies:
        - AWSLambdaBasicExecutionRole
  {{- end }}

  # Infrastructure Resources
  {{- if .HasFeature "dynamodb" }}
  UserTable:
    Type: AWS::DynamoDB::Table
    Properties:
      TableName: !Sub ${AWS::StackName}-users
      BillingMode: PAY_PER_REQUEST
      StreamSpecification:
        StreamViewType: NEW_AND_OLD_IMAGES
      AttributeDefinitions:
        - AttributeName: id
          AttributeType: S
        - AttributeName: email
          AttributeType: S
      KeySchema:
        - AttributeName: id
          KeyType: HASH
      GlobalSecondaryIndexes:
        - IndexName: email-index
          KeySchema:
            - AttributeName: email
              KeyType: HASH
          Projection:
            ProjectionType: ALL
      PointInTimeRecoverySpecification:
        PointInTimeRecoveryEnabled: true
      SSESpecification:
        SSEEnabled: true
      Tags:
        - Key: Application
          Value: {{.Name}}
        - Key: Environment
          Value: !Ref Environment
  {{- end }}

  {{- if .HasFeature "sqs" }}
  MessageQueue:
    Type: AWS::SQS::Queue
    Properties:
      QueueName: !Sub ${AWS::StackName}-messages
      VisibilityTimeout: 180
      MessageRetentionPeriod: 1209600
      RedrivePolicy:
        deadLetterTargetArn: !GetAtt DeadLetterQueue.Arn
        maxReceiveCount: 3
      KmsMasterKeyId: alias/aws/sqs
      Tags:
        - Key: Application
          Value: {{.Name}}
        - Key: Environment
          Value: !Ref Environment

  DeadLetterQueue:
    Type: AWS::SQS::Queue
    Properties:
      QueueName: !Sub ${AWS::StackName}-messages-dlq
      MessageRetentionPeriod: 1209600
      KmsMasterKeyId: alias/aws/sqs
      Tags:
        - Key: Application
          Value: {{.Name}}
        - Key: Environment
          Value: !Ref Environment
  {{- end }}

  {{- if .HasFeature "s3" }}
  StorageBucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: !Sub ${AWS::StackName}-storage-${AWS::AccountId}
      BucketEncryption:
        ServerSideEncryptionConfiguration:
          - ServerSideEncryptionByDefault:
              SSEAlgorithm: AES256
      VersioningConfiguration:
        Status: Enabled
      LifecycleConfiguration:
        Rules:
          - Id: DeleteOldVersions
            NoncurrentVersionExpirationInDays: 30
            Status: Enabled
      PublicAccessBlockConfiguration:
        BlockPublicAcls: true
        BlockPublicPolicy: true
        IgnorePublicAcls: true
        RestrictPublicBuckets: true
      Tags:
        - Key: Application
          Value: {{.Name}}
        - Key: Environment
          Value: !Ref Environment
  {{- end }}

  {{- if .HasFeature "cognito" }}
  CognitoUserPool:
    Type: AWS::Cognito::UserPool
    Properties:
      UserPoolName: !Sub ${AWS::StackName}-users
      UsernameAttributes:
        - email
      AutoVerifiedAttributes:
        - email
      PasswordPolicy:
        MinimumLength: 8
        RequireLowercase: true
        RequireNumbers: true
        RequireSymbols: true
        RequireUppercase: true
      MfaConfiguration: OPTIONAL
      EnabledMfas:
        - SOFTWARE_TOKEN_MFA
      UserPoolTags:
        Application: {{.Name}}
        Environment: !Ref Environment

  CognitoUserPoolClient:
    Type: AWS::Cognito::UserPoolClient
    Properties:
      ClientName: !Sub ${AWS::StackName}-client
      UserPoolId: !Ref CognitoUserPool
      GenerateSecret: false
      ExplicitAuthFlows:
        - ALLOW_USER_PASSWORD_AUTH
        - ALLOW_REFRESH_TOKEN_AUTH
  {{- end }}

Outputs:
  {{- if .HasFeature "api" }}
  ApiUrl:
    Description: API Gateway endpoint URL
    Value: !Sub https://${ApiGateway}.execute-api.${AWS::Region}.amazonaws.com/${Environment}
  {{- end }}

  {{- if .HasFeature "dynamodb" }}
  UserTableName:
    Description: DynamoDB table name for users
    Value: !Ref UserTable
  {{- end }}

  {{- if .HasFeature "sqs" }}
  MessageQueueUrl:
    Description: SQS queue URL
    Value: !Ref MessageQueue
  
  DeadLetterQueueUrl:
    Description: DLQ URL
    Value: !Ref DeadLetterQueue
  {{- end }}

  {{- if .HasFeature "s3" }}
  StorageBucketName:
    Description: S3 bucket name
    Value: !Ref StorageBucket
  {{- end }}

  {{- if .HasFeature "cognito" }}
  UserPoolId:
    Description: Cognito User Pool ID
    Value: !Ref CognitoUserPool
  
  UserPoolClientId:
    Description: Cognito User Pool Client ID
    Value: !Ref CognitoUserPoolClient
  {{- end }}
`

const SAMConfig = `version = 0.1

[default]
[default.global.parameters]
stack_name = "{{.Name}}"

[default.build.parameters]
cached = true
parallel = true

[default.deploy.parameters]
capabilities = "CAPABILITY_IAM"
confirm_changeset = true
resolve_s3 = true

[dev]
[dev.deploy.parameters]
stack_name = "{{.Name}}-dev"
s3_prefix = "{{.Name}}-dev"
region = "us-east-1"
confirm_changeset = false
capabilities = "CAPABILITY_IAM"
parameter_overrides = "Environment=dev LogLevel=debug"

[staging]
[staging.deploy.parameters]
stack_name = "{{.Name}}-staging"
s3_prefix = "{{.Name}}-staging"
region = "us-east-1"
confirm_changeset = true
capabilities = "CAPABILITY_IAM"
parameter_overrides = "Environment=staging LogLevel=info"

[prod]
[prod.deploy.parameters]
stack_name = "{{.Name}}-prod"
s3_prefix = "{{.Name}}-prod"
region = "us-east-1"
confirm_changeset = true
capabilities = "CAPABILITY_IAM"
parameter_overrides = "Environment=prod LogLevel=warn"
`

const BuildSpec = `version: 0.2

phases:
  pre_build:
    commands:
      - echo Installing dependencies...
      - go mod download
      
  build:
    commands:
      - echo Building Lambda functions...
      - make build
      
  post_build:
    commands:
      - echo Build completed on ` + "`date`" + `
      - sam package --s3-bucket $BUCKET_NAME --output-template-file packaged.yaml
      - sam deploy --template-file packaged.yaml --stack-name $STACK_NAME --capabilities CAPABILITY_IAM --no-confirm-changeset

artifacts:
  files:
    - packaged.yaml
    - build/**/*
`

const SAMEnvDev = `Environment=dev
LogLevel=debug
{{- if .HasFeature "api" }}
CorsOrigins=http://localhost:3000,http://localhost:8080
{{- end }}
{{- if .HasFeature "cognito" }}
CognitoUserPoolId=us-east-1_dev123456
CognitoClientId=dev1234567890abcdef
{{- end }}
`

const SAMEnvStaging = `Environment=staging
LogLevel=info
{{- if .HasFeature "api" }}
CorsOrigins=https://staging.{{.Name}}.com
{{- end }}
{{- if .HasFeature "cognito" }}
CognitoUserPoolId=us-east-1_stg123456
CognitoClientId=stg1234567890abcdef
{{- end }}
`

const SAMEnvProd = `Environment=prod
LogLevel=warn
{{- if .HasFeature "api" }}
CorsOrigins=https://{{.Name}}.com,https://www.{{.Name}}.com
{{- end }}
{{- if .HasFeature "cognito" }}
CognitoUserPoolId=us-east-1_prd123456
CognitoClientId=prd1234567890abcdef
{{- end }}
`

// CDK Templates
const CDKConfig = `{
  "app": "npx ts-node --prefer-ts-exts bin/app.ts",
  "watch": {
    "include": [
      "**"
    ],
    "exclude": [
      "README.md",
      "cdk*.json",
      "**/*.d.ts",
      "**/*.js",
      "tsconfig.json",
      "package*.json",
      "yarn.lock",
      "node_modules",
      "test"
    ]
  },
  "context": {
    "@aws-cdk/aws-apigateway:usagePlanKeyOrderInsensitiveId": true,
    "@aws-cdk/core:stackRelativeExports": true,
    "@aws-cdk/aws-lambda:recognizeVersionProps": true,
    "@aws-cdk/aws-lambda:recognizeLayerVersion": true,
    "@aws-cdk/core:checkSecretUsage": true,
    "@aws-cdk/core:target-partitions": [
      "aws",
      "aws-cn"
    ],
    "@aws-cdk-containers/ecs-service-extensions:enableDefaultLogDriver": true,
    "@aws-cdk/core:enablePartitionLiterals": true,
    "@aws-cdk/core:validateSnapshotRemovalPolicy": true,
    "@aws-cdk/aws-codepipeline:crossAccountKeyAliasStackSafeResourceName": true,
    "@aws-cdk/aws-s3:createDefaultLoggingPolicy": true,
    "@aws-cdk/aws-sns-subscriptions:restrictSqsDescryption": true,
    "@aws-cdk/aws-apigateway:disableCloudWatchRole": true,
    "@aws-cdk/core:enablePartitionLiterals": true
  }
}
`

const CDKTSConfig = `{
  "compilerOptions": {
    "target": "ES2020",
    "module": "commonjs",
    "lib": [
      "es2020"
    ],
    "declaration": true,
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "noImplicitThis": true,
    "alwaysStrict": true,
    "noUnusedLocals": false,
    "noUnusedParameters": false,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": false,
    "inlineSourceMap": true,
    "inlineSources": true,
    "experimentalDecorators": true,
    "strictPropertyInitialization": false,
    "typeRoots": [
      "./node_modules/@types"
    ]
  },
  "exclude": [
    "node_modules",
    "cdk.out"
  ]
}
`

const CDKPackageJSON = `{
  "name": "{{.Name}}-cdk",
  "version": "0.1.0",
  "bin": {
    "app": "bin/app.js"
  },
  "scripts": {
    "build": "tsc",
    "watch": "tsc -w",
    "test": "jest",
    "cdk": "cdk",
    "deploy:dev": "cdk deploy --all --context env=dev",
    "deploy:staging": "cdk deploy --all --context env=staging",
    "deploy:prod": "cdk deploy --all --context env=prod --require-approval broadening"
  },
  "devDependencies": {
    "@types/jest": "^29.5.5",
    "@types/node": "20.8.10",
    "jest": "^29.7.0",
    "ts-jest": "^29.1.1",
    "aws-cdk": "2.110.0",
    "ts-node": "^10.9.1",
    "typescript": "~5.2.2"
  },
  "dependencies": {
    "aws-cdk-lib": "2.110.0",
    "constructs": "^10.0.0",
    "source-map-support": "^0.5.21"
  }
}
`

const CDKStack = `import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
{{- if .HasFeature "dynamodb" }}
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
{{- end }}
{{- if .HasFeature "sqs" }}
import * as sqs from 'aws-cdk-lib/aws-sqs';
import { SqsEventSource } from 'aws-cdk-lib/aws-lambda-event-sources';
{{- end }}
{{- if .HasFeature "s3" }}
import * as s3 from 'aws-cdk-lib/aws-s3';
{{- end }}
{{- if .HasFeature "cognito" }}
import * as cognito from 'aws-cdk-lib/aws-cognito';
{{- end }}
import * as logs from 'aws-cdk-lib/aws-logs';
import * as path from 'path';

export interface {{.Name}}StackProps extends cdk.StackProps {
  environment: 'dev' | 'staging' | 'prod';
}

export class {{.Name}}Stack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: {{.Name}}StackProps) {
    super(scope, id, props);

    const env = props.environment;

    {{- if .HasFeature "dynamodb" }}
    // DynamoDB Table
    const userTable = new dynamodb.Table(this, 'UserTable', {
      tableName: ` + "`${this.stackName}-users`" + `,
      partitionKey: {
        name: 'id',
        type: dynamodb.AttributeType.STRING
      },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      encryption: dynamodb.TableEncryption.AWS_MANAGED,
      pointInTimeRecovery: true,
      stream: dynamodb.StreamViewType.NEW_AND_OLD_IMAGES,
    });

    userTable.addGlobalSecondaryIndex({
      indexName: 'email-index',
      partitionKey: {
        name: 'email',
        type: dynamodb.AttributeType.STRING
      },
      projectionType: dynamodb.ProjectionType.ALL
    });
    {{- end }}

    {{- if .HasFeature "sqs" }}
    // SQS Queues
    const deadLetterQueue = new sqs.Queue(this, 'DeadLetterQueue', {
      queueName: ` + "`${this.stackName}-messages-dlq`" + `,
      retentionPeriod: cdk.Duration.days(14),
      encryption: sqs.QueueEncryption.KMS_MANAGED,
    });

    const messageQueue = new sqs.Queue(this, 'MessageQueue', {
      queueName: ` + "`${this.stackName}-messages`" + `,
      visibilityTimeout: cdk.Duration.seconds(180),
      deadLetterQueue: {
        queue: deadLetterQueue,
        maxReceiveCount: 3,
      },
      encryption: sqs.QueueEncryption.KMS_MANAGED,
    });
    {{- end }}

    {{- if .HasFeature "s3" }}
    // S3 Bucket
    const storageBucket = new s3.Bucket(this, 'StorageBucket', {
      bucketName: ` + "`${this.stackName}-storage-${this.account}`" + `,
      encryption: s3.BucketEncryption.S3_MANAGED,
      versioned: true,
      lifecycleRules: [{
        noncurrentVersionExpiration: cdk.Duration.days(30),
      }],
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      removalPolicy: env === 'prod' ? cdk.RemovalPolicy.RETAIN : cdk.RemovalPolicy.DESTROY,
      autoDeleteObjects: env !== 'prod',
    });
    {{- end }}

    {{- if .HasFeature "cognito" }}
    // Cognito User Pool
    const userPool = new cognito.UserPool(this, 'UserPool', {
      userPoolName: ` + "`${this.stackName}-users`" + `,
      selfSignUpEnabled: true,
      signInAliases: {
        email: true,
      },
      autoVerify: {
        email: true,
      },
      passwordPolicy: {
        minLength: 8,
        requireLowercase: true,
        requireUppercase: true,
        requireDigits: true,
        requireSymbols: true,
      },
      mfa: cognito.Mfa.OPTIONAL,
      mfaSecondFactor: {
        sms: false,
        otp: true,
      },
      removalPolicy: env === 'prod' ? cdk.RemovalPolicy.RETAIN : cdk.RemovalPolicy.DESTROY,
    });

    const userPoolClient = new cognito.UserPoolClient(this, 'UserPoolClient', {
      userPool,
      generateSecret: false,
      authFlows: {
        userPassword: true,
        custom: true,
      },
    });
    {{- end }}

    // Lambda Functions
    const lambdaEnvironment = {
      APP_NAME: '{{.Name}}',
      APP_ENV: env,
      LOG_LEVEL: env === 'dev' ? 'debug' : env === 'staging' ? 'info' : 'warn',
      {{- if .HasFeature "dynamodb" }}
      DYNAMODB_TABLE_NAME: userTable.tableName,
      {{- end }}
      {{- if .HasFeature "sqs" }}
      SQS_QUEUE_URL: messageQueue.queueUrl,
      {{- end }}
      {{- if .HasFeature "s3" }}
      S3_BUCKET_NAME: storageBucket.bucketName,
      {{- end }}
      {{- if .HasFeature "cognito" }}
      COGNITO_USER_POOL_ID: userPool.userPoolId,
      COGNITO_CLIENT_ID: userPoolClient.userPoolClientId,
      {{- end }}
    };

    {{- if eq .Architecture "clean" }}
    const userFunction = new lambda.Function(this, 'UserFunction', {
      functionName: ` + "`${this.stackName}-user-handler`" + `,
      runtime: lambda.Runtime.PROVIDED_AL2023,
      handler: 'bootstrap',
      code: lambda.Code.fromAsset(path.join(__dirname, '../../build/user')),
      memorySize: 512,
      timeout: cdk.Duration.seconds(30),
      environment: lambdaEnvironment,
      tracing: lambda.Tracing.ACTIVE,
      logRetention: logs.RetentionDays.ONE_WEEK,
    });

    {{- if .HasFeature "dynamodb" }}
    userTable.grantReadWriteData(userFunction);
    {{- end }}
    {{- end }}

    {{- if .HasFeature "sqs" }}
    const messageProcessorFunction = new lambda.Function(this, 'MessageProcessorFunction', {
      functionName: ` + "`${this.stackName}-message-processor`" + `,
      runtime: lambda.Runtime.PROVIDED_AL2023,
      handler: 'bootstrap',
      code: lambda.Code.fromAsset(path.join(__dirname, '../../build/message-processor')),
      memorySize: 512,
      timeout: cdk.Duration.seconds(180),
      environment: lambdaEnvironment,
      tracing: lambda.Tracing.ACTIVE,
      logRetention: logs.RetentionDays.ONE_WEEK,
    });

    messageQueue.grantConsumeMessages(messageProcessorFunction);
    messageProcessorFunction.addEventSource(new SqsEventSource(messageQueue, {
      batchSize: 10,
      maxBatchingWindow: cdk.Duration.seconds(5),
    }));
    {{- end }}

    {{- if .HasFeature "api" }}
    // API Gateway
    const api = new apigateway.RestApi(this, 'Api', {
      restApiName: ` + "`${this.stackName}-api`" + `,
      deployOptions: {
        stageName: env,
        tracingEnabled: true,
        loggingLevel: apigateway.MethodLoggingLevel.INFO,
        dataTraceEnabled: env === 'dev',
        metricsEnabled: true,
      },
      defaultCorsPreflightOptions: {
        allowOrigins: env === 'dev' 
          ? ['http://localhost:3000', 'http://localhost:8080']
          : env === 'staging'
          ? ['https://staging.{{.Name}}.com']
          : ['https://{{.Name}}.com', 'https://www.{{.Name}}.com'],
        allowMethods: apigateway.Cors.ALL_METHODS,
        allowHeaders: ['Content-Type', 'X-Amz-Date', 'Authorization', 'X-Api-Key', 'X-Request-ID'],
      },
    });

    {{- if eq .Architecture "clean" }}
    // User endpoints
    const users = api.root.addResource('users');
    const userIntegration = new apigateway.LambdaIntegration(userFunction);
    
    users.addMethod('POST', userIntegration);
    users.addMethod('GET', userIntegration);
    
    const userById = users.addResource('{id}');
    userById.addMethod('GET', userIntegration);
    userById.addMethod('PUT', userIntegration);
    userById.addMethod('DELETE', userIntegration);
    {{- end }}

    // Output the API URL
    new cdk.CfnOutput(this, 'ApiUrl', {
      value: api.url,
      description: 'API Gateway endpoint URL',
    });
    {{- end }}

    // Stack outputs
    {{- if .HasFeature "dynamodb" }}
    new cdk.CfnOutput(this, 'UserTableName', {
      value: userTable.tableName,
      description: 'DynamoDB table name for users',
    });
    {{- end }}

    {{- if .HasFeature "sqs" }}
    new cdk.CfnOutput(this, 'MessageQueueUrl', {
      value: messageQueue.queueUrl,
      description: 'SQS queue URL',
    });
    {{- end }}

    {{- if .HasFeature "cognito" }}
    new cdk.CfnOutput(this, 'UserPoolId', {
      value: userPool.userPoolId,
      description: 'Cognito User Pool ID',
    });
    
    new cdk.CfnOutput(this, 'UserPoolClientId', {
      value: userPoolClient.userPoolClientId,
      description: 'Cognito User Pool Client ID',
    });
    {{- end }}
  }
}
`

const CDKApp = `#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { {{.Name}}Stack } from '../lib/stack';

const app = new cdk.App();

const env = app.node.tryGetContext('env') || 'dev';

new {{.Name}}Stack(app, ` + "`{{.Name}}Stack-${env}`" + `, {
  environment: env,
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: process.env.CDK_DEFAULT_REGION || 'us-east-1',
  },
  tags: {
    Application: '{{.Name}}',
    Environment: env,
  },
});
`

const CDKTest = `import * as cdk from 'aws-cdk-lib';
import { Template } from 'aws-cdk-lib/assertions';
import { {{.Name}}Stack } from '../lib/stack';

describe('{{.Name}}Stack', () => {
  test('Stack creates required resources', () => {
    const app = new cdk.App();
    const stack = new {{.Name}}Stack(app, 'TestStack', {
      environment: 'dev',
    });
    
    const template = Template.fromStack(stack);

    {{- if .HasFeature "api" }}
    // Check API Gateway exists
    template.hasResourceProperties('AWS::ApiGateway::RestApi', {
      Name: 'TestStack-api',
    });
    {{- end }}

    {{- if .HasFeature "dynamodb" }}
    // Check DynamoDB table exists
    template.hasResourceProperties('AWS::DynamoDB::Table', {
      TableName: 'TestStack-users',
      BillingMode: 'PAY_PER_REQUEST',
    });
    {{- end }}

    {{- if .HasFeature "sqs" }}
    // Check SQS queue exists
    template.hasResourceProperties('AWS::SQS::Queue', {
      QueueName: 'TestStack-messages',
    });
    {{- end }}

    {{- if .HasFeature "cognito" }}
    // Check Cognito User Pool exists
    template.hasResourceProperties('AWS::Cognito::UserPool', {
      UserPoolName: 'TestStack-users',
    });
    {{- end }}
  });
});
`

// Serverless Framework templates
const ServerlessYML = `service: {{.Name}}

frameworkVersion: '3'

provider:
  name: aws
  runtime: provided.al2023
  architecture: x86_64
  stage: ${opt:stage, 'dev'}
  region: ${opt:region, 'us-east-1'}
  memorySize: 512
  timeout: 30
  tracing:
    lambda: true
    apiGateway: true
  environment:
    APP_NAME: ${self:service}
    APP_ENV: ${self:provider.stage}
    LOG_LEVEL: ${self:custom.logLevel.${self:provider.stage}}
    {{- if .HasFeature "dynamodb" }}
    DYNAMODB_TABLE_NAME: ${self:service}-${self:provider.stage}-users
    {{- end }}
    {{- if .HasFeature "sqs" }}
    SQS_QUEUE_URL: !Ref MessageQueue
    {{- end }}
  iam:
    role:
      statements:
        {{- if .HasFeature "dynamodb" }}
        - Effect: Allow
          Action:
            - dynamodb:DescribeTable
            - dynamodb:Query
            - dynamodb:Scan
            - dynamodb:GetItem
            - dynamodb:PutItem
            - dynamodb:UpdateItem
            - dynamodb:DeleteItem
          Resource:
            - !GetAtt UserTable.Arn
            - !Sub "${UserTable.Arn}/index/*"
        {{- end }}
        {{- if .HasFeature "sqs" }}
        - Effect: Allow
          Action:
            - sqs:SendMessage
            - sqs:ReceiveMessage
            - sqs:DeleteMessage
            - sqs:GetQueueAttributes
          Resource:
            - !GetAtt MessageQueue.Arn
        {{- end }}
        {{- if .HasFeature "s3" }}
        - Effect: Allow
          Action:
            - s3:GetObject
            - s3:PutObject
            - s3:DeleteObject
          Resource:
            - !Sub "${StorageBucket.Arn}/*"
        {{- end }}

custom:
  logLevel:
    dev: debug
    staging: info
    prod: warn
  {{- if .HasFeature "api" }}
  cors:
    dev:
      origins:
        - http://localhost:3000
        - http://localhost:8080
    staging:
      origins:
        - https://staging.{{.Name}}.com
    prod:
      origins:
        - https://{{.Name}}.com
        - https://www.{{.Name}}.com
  {{- end }}

functions:
  {{- if eq .Architecture "clean" }}
  userHandler:
    handler: bootstrap
    package:
      artifact: build/user.zip
    {{- if .HasFeature "api" }}
    events:
      - http:
          path: users
          method: POST
          cors: ${self:custom.cors.${self:provider.stage}}
      - http:
          path: users
          method: GET
          cors: ${self:custom.cors.${self:provider.stage}}
      - http:
          path: users/{id}
          method: GET
          cors: ${self:custom.cors.${self:provider.stage}}
      - http:
          path: users/{id}
          method: PUT
          cors: ${self:custom.cors.${self:provider.stage}}
      - http:
          path: users/{id}
          method: DELETE
          cors: ${self:custom.cors.${self:provider.stage}}
    {{- end }}
  {{- end }}

  {{- if .HasFeature "sqs" }}
  messageProcessor:
    handler: bootstrap
    package:
      artifact: build/message-processor.zip
    events:
      - sqs:
          arn: !GetAtt MessageQueue.Arn
          batchSize: 10
  {{- end }}

resources:
  Resources:
    {{- if .HasFeature "dynamodb" }}
    UserTable:
      Type: AWS::DynamoDB::Table
      Properties:
        TableName: ${self:service}-${self:provider.stage}-users
        BillingMode: PAY_PER_REQUEST
        StreamSpecification:
          StreamViewType: NEW_AND_OLD_IMAGES
        AttributeDefinitions:
          - AttributeName: id
            AttributeType: S
          - AttributeName: email
            AttributeType: S
        KeySchema:
          - AttributeName: id
            KeyType: HASH
        GlobalSecondaryIndexes:
          - IndexName: email-index
            KeySchema:
              - AttributeName: email
                KeyType: HASH
            Projection:
              ProjectionType: ALL
        PointInTimeRecoverySpecification:
          PointInTimeRecoveryEnabled: true
        SSESpecification:
          SSEEnabled: true
    {{- end }}

    {{- if .HasFeature "sqs" }}
    MessageQueue:
      Type: AWS::SQS::Queue
      Properties:
        QueueName: ${self:service}-${self:provider.stage}-messages
        VisibilityTimeout: 180
        RedrivePolicy:
          deadLetterTargetArn: !GetAtt DeadLetterQueue.Arn
          maxReceiveCount: 3
        KmsMasterKeyId: alias/aws/sqs

    DeadLetterQueue:
      Type: AWS::SQS::Queue
      Properties:
        QueueName: ${self:service}-${self:provider.stage}-messages-dlq
        MessageRetentionPeriod: 1209600
        KmsMasterKeyId: alias/aws/sqs
    {{- end }}

    {{- if .HasFeature "s3" }}
    StorageBucket:
      Type: AWS::S3::Bucket
      Properties:
        BucketName: ${self:service}-${self:provider.stage}-storage-${aws:accountId}
        BucketEncryption:
          ServerSideEncryptionConfiguration:
            - ServerSideEncryptionByDefault:
                SSEAlgorithm: AES256
        VersioningConfiguration:
          Status: Enabled
        PublicAccessBlockConfiguration:
          BlockPublicAcls: true
          BlockPublicPolicy: true
          IgnorePublicAcls: true
          RestrictPublicBuckets: true
    {{- end }}

    {{- if .HasFeature "cognito" }}
    CognitoUserPool:
      Type: AWS::Cognito::UserPool
      Properties:
        UserPoolName: ${self:service}-${self:provider.stage}-users
        UsernameAttributes:
          - email
        AutoVerifiedAttributes:
          - email
        PasswordPolicy:
          MinimumLength: 8
          RequireLowercase: true
          RequireNumbers: true
          RequireSymbols: true
          RequireUppercase: true

    CognitoUserPoolClient:
      Type: AWS::Cognito::UserPoolClient
      Properties:
        ClientName: ${self:service}-${self:provider.stage}-client
        UserPoolId: !Ref CognitoUserPool
        GenerateSecret: false
        ExplicitAuthFlows:
          - ALLOW_USER_PASSWORD_AUTH
          - ALLOW_REFRESH_TOKEN_AUTH
    {{- end }}

  Outputs:
    {{- if .HasFeature "api" }}
    ApiUrl:
      Value: !Sub https://${ApiGatewayRestApi}.execute-api.${AWS::Region}.amazonaws.com/${self:provider.stage}
    {{- end }}
    {{- if .HasFeature "dynamodb" }}
    UserTableName:
      Value: !Ref UserTable
    {{- end }}
    {{- if .HasFeature "sqs" }}
    MessageQueueUrl:
      Value: !Ref MessageQueue
    {{- end }}
    {{- if .HasFeature "cognito" }}
    UserPoolId:
      Value: !Ref CognitoUserPool
    UserPoolClientId:
      Value: !Ref CognitoUserPoolClient
    {{- end }}

plugins:
  - serverless-offline
  - serverless-plugin-tracing
`

const ServerlessEnv = `# Serverless environment configuration
logLevel: ${self:custom.logLevel.${self:provider.stage}}
`

const ServerlessEnvDev = `# Development environment
apiCorsOrigins:
  - http://localhost:3000
  - http://localhost:8080
`

const ServerlessEnvStaging = `# Staging environment
apiCorsOrigins:
  - https://staging.{{.Name}}.com
`

const ServerlessEnvProd = `# Production environment
apiCorsOrigins:
  - https://{{.Name}}.com
  - https://www.{{.Name}}.com
`

// Terraform templates
const TerraformMain = `terraform {
  required_version = ">= 1.5"
  
  backend "s3" {
    # Configure your backend
    # bucket = "your-terraform-state-bucket"
    # key    = "{{.Name}}/terraform.tfstate"
    # region = "us-east-1"
  }
}

provider "aws" {
  region = var.aws_region
  
  default_tags {
    tags = {
      Application = var.app_name
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}

locals {
  app_prefix = "${var.app_name}-${var.environment}"
}

{{- if .HasFeature "api" }}
# API Gateway
resource "aws_api_gateway_rest_api" "api" {
  name        = "${local.app_prefix}-api"
  description = "API Gateway for ${var.app_name}"
  
  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

resource "aws_api_gateway_deployment" "api" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  
  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_rest_api.api.root_resource_id,
      # Add other resources that should trigger redeployment
    ]))
  }
  
  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "api" {
  deployment_id = aws_api_gateway_deployment.api.id
  rest_api_id   = aws_api_gateway_rest_api.api.id
  stage_name    = var.environment
  
  xray_tracing_enabled = true
  
  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gateway.arn
    format = jsonencode({
      requestId      = "$context.requestId"
      ip             = "$context.identity.sourceIp"
      requestTime    = "$context.requestTime"
      httpMethod     = "$context.httpMethod"
      routeKey       = "$context.routeKey"
      status         = "$context.status"
      protocol       = "$context.protocol"
      responseLength = "$context.responseLength"
    })
  }
}

resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigateway/${local.app_prefix}"
  retention_in_days = var.log_retention_days
}
{{- end }}

{{- if .HasFeature "dynamodb" }}
# DynamoDB Table
resource "aws_dynamodb_table" "users" {
  name         = "${local.app_prefix}-users"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"
  
  attribute {
    name = "id"
    type = "S"
  }
  
  attribute {
    name = "email"
    type = "S"
  }
  
  global_secondary_index {
    name            = "email-index"
    hash_key        = "email"
    projection_type = "ALL"
  }
  
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"
  
  point_in_time_recovery {
    enabled = true
  }
  
  server_side_encryption {
    enabled = true
  }
  
  lifecycle {
    prevent_destroy = true
  }
}
{{- end }}

{{- if .HasFeature "sqs" }}
# SQS Queues
resource "aws_sqs_queue" "dlq" {
  name                      = "${local.app_prefix}-messages-dlq"
  message_retention_seconds = 1209600 # 14 days
  kms_master_key_id        = "alias/aws/sqs"
}

resource "aws_sqs_queue" "messages" {
  name                      = "${local.app_prefix}-messages"
  visibility_timeout_seconds = 180
  message_retention_seconds = 1209600
  kms_master_key_id        = "alias/aws/sqs"
  
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq.arn
    maxReceiveCount     = 3
  })
}
{{- end }}

{{- if .HasFeature "s3" }}
# S3 Bucket
resource "aws_s3_bucket" "storage" {
  bucket = "${local.app_prefix}-storage-${data.aws_caller_identity.current.account_id}"
}

resource "aws_s3_bucket_versioning" "storage" {
  bucket = aws_s3_bucket.storage.id
  
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "storage" {
  bucket = aws_s3_bucket.storage.id
  
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "storage" {
  bucket = aws_s3_bucket.storage.id
  
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "storage" {
  bucket = aws_s3_bucket.storage.id
  
  rule {
    id     = "delete-old-versions"
    status = "Enabled"
    
    noncurrent_version_expiration {
      noncurrent_days = 30
    }
  }
}
{{- end }}

{{- if .HasFeature "cognito" }}
# Cognito User Pool
resource "aws_cognito_user_pool" "users" {
  name = "${local.app_prefix}-users"
  
  username_attributes      = ["email"]
  auto_verified_attributes = ["email"]
  
  password_policy {
    minimum_length    = 8
    require_lowercase = true
    require_numbers   = true
    require_symbols   = true
    require_uppercase = true
  }
  
  mfa_configuration = "OPTIONAL"
  
  software_token_mfa_configuration {
    enabled = true
  }
  
  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
  }
  
  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_cognito_user_pool_client" "client" {
  name         = "${local.app_prefix}-client"
  user_pool_id = aws_cognito_user_pool.users.id
  
  generate_secret = false
  
  explicit_auth_flows = [
    "ALLOW_USER_PASSWORD_AUTH",
    "ALLOW_REFRESH_TOKEN_AUTH"
  ]
}
{{- end }}

# Lambda Functions
module "user_function" {
  source = "./modules/lambda"
  
  function_name = "${local.app_prefix}-user-handler"
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  filename      = "../build/user.zip"
  
  environment_variables = {
    APP_NAME     = var.app_name
    APP_ENV      = var.environment
    LOG_LEVEL    = var.log_level
    {{- if .HasFeature "dynamodb" }}
    DYNAMODB_TABLE_NAME = aws_dynamodb_table.users.name
    {{- end }}
  }
  
  {{- if .HasFeature "dynamodb" }}
  attach_policy_statements = true
  policy_statements = {
    dynamodb = {
      effect = "Allow"
      actions = [
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:UpdateItem",
        "dynamodb:DeleteItem",
        "dynamodb:Query",
        "dynamodb:Scan"
      ]
      resources = [
        aws_dynamodb_table.users.arn,
        "${aws_dynamodb_table.users.arn}/index/*"
      ]
    }
  }
  {{- end }}
}

{{- if .HasFeature "sqs" }}
module "message_processor_function" {
  source = "./modules/lambda"
  
  function_name = "${local.app_prefix}-message-processor"
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  filename      = "../build/message-processor.zip"
  timeout       = 180
  
  environment_variables = {
    APP_NAME      = var.app_name
    APP_ENV       = var.environment
    LOG_LEVEL     = var.log_level
    SQS_QUEUE_URL = aws_sqs_queue.messages.url
  }
  
  attach_policy_statements = true
  policy_statements = {
    sqs = {
      effect = "Allow"
      actions = [
        "sqs:ReceiveMessage",
        "sqs:DeleteMessage",
        "sqs:GetQueueAttributes"
      ]
      resources = [aws_sqs_queue.messages.arn]
    }
  }
}

# SQS trigger for Lambda
resource "aws_lambda_event_source_mapping" "sqs" {
  event_source_arn = aws_sqs_queue.messages.arn
  function_name    = module.message_processor_function.function_name
  batch_size       = 10
}
{{- end }}

# Data sources
data "aws_caller_identity" "current" {}
`

const TerraformVariables = `variable "app_name" {
  description = "Application name"
  type        = string
  default     = "{{.Name}}"
}

variable "environment" {
  description = "Environment (dev/staging/prod)"
  type        = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "log_level" {
  description = "Application log level"
  type        = string
  default     = "info"
  validation {
    condition     = contains(["debug", "info", "warn", "error"], var.log_level)
    error_message = "Log level must be debug, info, warn, or error."
  }
}

variable "log_retention_days" {
  description = "CloudWatch log retention in days"
  type        = number
  default     = 7
}

{{- if .HasFeature "api" }}
variable "cors_origins" {
  description = "CORS allowed origins"
  type        = list(string)
  default     = ["*"]
}
{{- end }}
`

const TerraformOutputs = `{{- if .HasFeature "api" }}
output "api_url" {
  description = "API Gateway URL"
  value       = "${aws_api_gateway_stage.api.invoke_url}/"
}
{{- end }}

{{- if .HasFeature "dynamodb" }}
output "user_table_name" {
  description = "DynamoDB table name"
  value       = aws_dynamodb_table.users.name
}

output "user_table_arn" {
  description = "DynamoDB table ARN"
  value       = aws_dynamodb_table.users.arn
}
{{- end }}

{{- if .HasFeature "sqs" }}
output "message_queue_url" {
  description = "SQS queue URL"
  value       = aws_sqs_queue.messages.url
}

output "dlq_url" {
  description = "Dead letter queue URL"
  value       = aws_sqs_queue.dlq.url
}
{{- end }}

{{- if .HasFeature "s3" }}
output "storage_bucket_name" {
  description = "S3 bucket name"
  value       = aws_s3_bucket.storage.id
}
{{- end }}

{{- if .HasFeature "cognito" }}
output "user_pool_id" {
  description = "Cognito User Pool ID"
  value       = aws_cognito_user_pool.users.id
}

output "user_pool_client_id" {
  description = "Cognito User Pool Client ID"
  value       = aws_cognito_user_pool_client.client.id
}
{{- end }}

output "lambda_function_names" {
  description = "Lambda function names"
  value = {
    user_handler = module.user_function.function_name
    {{- if .HasFeature "sqs" }}
    message_processor = module.message_processor_function.function_name
    {{- end }}
  }
}
`

const TerraformVersions = `terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
`

const TerraformEnvDev = `environment        = "dev"
log_level          = "debug"
log_retention_days = 3
{{- if .HasFeature "api" }}
cors_origins = ["http://localhost:3000", "http://localhost:8080"]
{{- end }}
`

const TerraformEnvProd = `environment        = "prod"
log_level          = "warn"
log_retention_days = 30
{{- if .HasFeature "api" }}
cors_origins = ["https://{{.Name}}.com", "https://www.{{.Name}}.com"]
{{- end }}
`

const TerraformLambdaModule = `variable "function_name" {
  description = "Lambda function name"
  type        = string
}

variable "handler" {
  description = "Lambda function handler"
  type        = string
}

variable "runtime" {
  description = "Lambda runtime"
  type        = string
}

variable "filename" {
  description = "Path to the function's deployment package"
  type        = string
}

variable "memory_size" {
  description = "Amount of memory in MB"
  type        = number
  default     = 512
}

variable "timeout" {
  description = "Function timeout in seconds"
  type        = number
  default     = 30
}

variable "environment_variables" {
  description = "Environment variables"
  type        = map(string)
  default     = {}
}

variable "attach_policy_statements" {
  description = "Whether to attach policy statements"
  type        = bool
  default     = false
}

variable "policy_statements" {
  description = "Map of policy statements"
  type        = any
  default     = {}
}

resource "aws_lambda_function" "this" {
  function_name = var.function_name
  role          = aws_iam_role.lambda.arn
  handler       = var.handler
  runtime       = var.runtime
  memory_size   = var.memory_size
  timeout       = var.timeout
  filename      = var.filename
  
  environment {
    variables = var.environment_variables
  }
  
  tracing_config {
    mode = "Active"
  }
}

resource "aws_iam_role" "lambda" {
  name = "${var.function_name}-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_basic" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.lambda.name
}

resource "aws_iam_role_policy_attachment" "lambda_xray" {
  policy_arn = "arn:aws:iam::aws:policy/AWSXRayDaemonWriteAccess"
  role       = aws_iam_role.lambda.name
}

resource "aws_iam_role_policy" "lambda" {
  count = var.attach_policy_statements ? 1 : 0
  name  = "${var.function_name}-policy"
  role  = aws_iam_role.lambda.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [for k, v in var.policy_statements : {
      Effect   = v.effect
      Action   = v.actions
      Resource = v.resources
    }]
  })
}

resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${var.function_name}"
  retention_in_days = 7
}

output "function_name" {
  value = aws_lambda_function.this.function_name
}

output "function_arn" {
  value = aws_lambda_function.this.arn
}

output "invoke_arn" {
  value = aws_lambda_function.this.invoke_arn
}
`