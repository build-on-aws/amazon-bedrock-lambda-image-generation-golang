package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdkapigatewayv2alpha/v2"
	"github.com/aws/aws-cdk-go/awscdkapigatewayv2integrationsalpha/v2"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const functionDir = "../function"

type BedrockLambdaImgeGenWebsiteStackProps struct {
	awscdk.StackProps
}

func NewBedrockLambdaImgeGenWebsiteStack(scope constructs.Construct, id string, props *BedrockLambdaImgeGenWebsiteStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	apigw := awscdkapigatewayv2alpha.NewHttpApi(stack, jsii.String("image-gen-http-api"), nil)

	bucket := awss3.NewBucket(stack, jsii.String("website-s3-bucket"), &awss3.BucketProps{
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
	})

	oai := awscloudfront.NewOriginAccessIdentity(stack, jsii.String("OAI"), nil)

	bucket.GrantRead(oai.GrantPrincipal(), "*")

	distribution := awscloudfront.NewDistribution(stack, jsii.String("MyDistribution"), &awscloudfront.DistributionProps{
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin: awscloudfrontorigins.NewS3Origin(bucket, &awscloudfrontorigins.S3OriginProps{
				OriginAccessIdentity: oai,
			}),
		},
		DefaultRootObject: jsii.String("index.html"), //name of the file in S3
	})

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("bedrock-imagegen-s3"),
		&awscdklambdagoalpha.GoFunctionProps{
			Runtime: awslambda.Runtime_GO_1_X(),
			Entry:   jsii.String(functionDir),
			Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
		})

	function.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:   jsii.Strings("bedrock:*"),
		Effect:    awsiam.Effect_ALLOW,
		Resources: jsii.Strings("*"),
	}))

	functionIntg := awscdkapigatewayv2integrationsalpha.NewHttpLambdaIntegration(jsii.String("function-integration"), function, nil)

	apigw.AddRoutes(&awscdkapigatewayv2alpha.AddRoutesOptions{
		Path:        jsii.String("/"),
		Methods:     &[]awscdkapigatewayv2alpha.HttpMethod{awscdkapigatewayv2alpha.HttpMethod_POST},
		Integration: functionIntg})

	awscdk.NewCfnOutput(stack, jsii.String("apigw URL"), &awscdk.CfnOutputProps{Value: apigw.Url(), Description: jsii.String("API Gateway endpoint")})

	awscdk.NewCfnOutput(stack, jsii.String("cloud front domain name"), &awscdk.CfnOutputProps{Value: distribution.DomainName(), Description: jsii.String("cloud front domain name")})

	awscdk.NewCfnOutput(stack, jsii.String("s3 bucket name"), &awscdk.CfnOutputProps{Value: bucket.BucketName(), Description: jsii.String("s3 bucket name")})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewBedrockLambdaImgeGenWebsiteStack(app, "BedrockLambdaImgeGenWebsiteStack", &BedrockLambdaImgeGenWebsiteStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return nil
}
