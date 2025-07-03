"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ObsidianBuildTrigger = void 0;
const cdk = require("aws-cdk-lib");
const lambda = require("aws-cdk-lib/aws-lambda");
const iam = require("aws-cdk-lib/aws-iam");
const constructs_1 = require("constructs");
class ObsidianBuildTrigger extends constructs_1.Construct {
    constructor(scope, id, props) {
        super(scope, id);
        // Create Lambda function to trigger builds
        this.triggerFunction = new lambda.Function(this, 'BuildTriggerFunction', {
            runtime: lambda.Runtime.NODEJS_18_X,
            handler: 'index.handler',
            code: lambda.Code.fromInline(`
        const AWS = require('aws-sdk');
        const codebuild = new AWS.CodeBuild();
        
        exports.handler = async (event) => {
          const params = {
            projectName: '${props.buildProject.projectName}',
            sourceVersion: event.branch || 'main'
          };
          
          try {
            const result = await codebuild.startBuild(params).promise();
            return {
              statusCode: 200,
              body: JSON.stringify({
                message: 'Build started successfully',
                buildId: result.build.id,
                buildArn: result.build.arn
              })
            };
          } catch (error) {
            return {
              statusCode: 500,
              body: JSON.stringify({
                message: 'Failed to start build',
                error: error.message
              })
            };
          }
        };
      `),
            timeout: cdk.Duration.seconds(30),
            functionName: 'obsidian-build-trigger',
        });
        // Grant permissions to trigger builds
        this.triggerFunction.addToRolePolicy(new iam.PolicyStatement({
            effect: iam.Effect.ALLOW,
            actions: [
                'codebuild:StartBuild',
                'codebuild:BatchGetBuilds',
            ],
            resources: [props.buildProject.projectArn],
        }));
        // Output the function name for easy invocation
        new cdk.CfnOutput(this, 'BuildTriggerFunctionName', {
            value: this.triggerFunction.functionName,
            description: 'Lambda function to trigger builds',
        });
        // Output command to trigger build
        new cdk.CfnOutput(this, 'TriggerBuildCommand', {
            value: `aws lambda invoke --function-name ${this.triggerFunction.functionName} --payload '{"branch":"main"}' response.json`,
            description: 'Command to trigger a build',
        });
    }
}
exports.ObsidianBuildTrigger = ObsidianBuildTrigger;
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiYnVpbGQtdHJpZ2dlci5qcyIsInNvdXJjZVJvb3QiOiIiLCJzb3VyY2VzIjpbImJ1aWxkLXRyaWdnZXIudHMiXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6Ijs7O0FBQUEsbUNBQW1DO0FBQ25DLGlEQUFpRDtBQUVqRCwyQ0FBMkM7QUFDM0MsMkNBQXVDO0FBTXZDLE1BQWEsb0JBQXFCLFNBQVEsc0JBQVM7SUFHakQsWUFBWSxLQUFnQixFQUFFLEVBQVUsRUFBRSxLQUF3QjtRQUNoRSxLQUFLLENBQUMsS0FBSyxFQUFFLEVBQUUsQ0FBQyxDQUFDO1FBRWpCLDJDQUEyQztRQUMzQyxJQUFJLENBQUMsZUFBZSxHQUFHLElBQUksTUFBTSxDQUFDLFFBQVEsQ0FBQyxJQUFJLEVBQUUsc0JBQXNCLEVBQUU7WUFDdkUsT0FBTyxFQUFFLE1BQU0sQ0FBQyxPQUFPLENBQUMsV0FBVztZQUNuQyxPQUFPLEVBQUUsZUFBZTtZQUN4QixJQUFJLEVBQUUsTUFBTSxDQUFDLElBQUksQ0FBQyxVQUFVLENBQUM7Ozs7Ozs0QkFNUCxLQUFLLENBQUMsWUFBWSxDQUFDLFdBQVc7Ozs7Ozs7Ozs7Ozs7Ozs7Ozs7Ozs7OztPQXdCbkQsQ0FBQztZQUNGLE9BQU8sRUFBRSxHQUFHLENBQUMsUUFBUSxDQUFDLE9BQU8sQ0FBQyxFQUFFLENBQUM7WUFDakMsWUFBWSxFQUFFLHdCQUF3QjtTQUN2QyxDQUFDLENBQUM7UUFFSCxzQ0FBc0M7UUFDdEMsSUFBSSxDQUFDLGVBQWUsQ0FBQyxlQUFlLENBQUMsSUFBSSxHQUFHLENBQUMsZUFBZSxDQUFDO1lBQzNELE1BQU0sRUFBRSxHQUFHLENBQUMsTUFBTSxDQUFDLEtBQUs7WUFDeEIsT0FBTyxFQUFFO2dCQUNQLHNCQUFzQjtnQkFDdEIsMEJBQTBCO2FBQzNCO1lBQ0QsU0FBUyxFQUFFLENBQUMsS0FBSyxDQUFDLFlBQVksQ0FBQyxVQUFVLENBQUM7U0FDM0MsQ0FBQyxDQUFDLENBQUM7UUFFSiwrQ0FBK0M7UUFDL0MsSUFBSSxHQUFHLENBQUMsU0FBUyxDQUFDLElBQUksRUFBRSwwQkFBMEIsRUFBRTtZQUNsRCxLQUFLLEVBQUUsSUFBSSxDQUFDLGVBQWUsQ0FBQyxZQUFZO1lBQ3hDLFdBQVcsRUFBRSxtQ0FBbUM7U0FDakQsQ0FBQyxDQUFDO1FBRUgsa0NBQWtDO1FBQ2xDLElBQUksR0FBRyxDQUFDLFNBQVMsQ0FBQyxJQUFJLEVBQUUscUJBQXFCLEVBQUU7WUFDN0MsS0FBSyxFQUFFLHFDQUFxQyxJQUFJLENBQUMsZUFBZSxDQUFDLFlBQVksOENBQThDO1lBQzNILFdBQVcsRUFBRSw0QkFBNEI7U0FDMUMsQ0FBQyxDQUFDO0lBQ0wsQ0FBQztDQUNGO0FBbkVELG9EQW1FQyIsInNvdXJjZXNDb250ZW50IjpbImltcG9ydCAqIGFzIGNkayBmcm9tICdhd3MtY2RrLWxpYic7XG5pbXBvcnQgKiBhcyBsYW1iZGEgZnJvbSAnYXdzLWNkay1saWIvYXdzLWxhbWJkYSc7XG5pbXBvcnQgKiBhcyBjb2RlYnVpbGQgZnJvbSAnYXdzLWNkay1saWIvYXdzLWNvZGVidWlsZCc7XG5pbXBvcnQgKiBhcyBpYW0gZnJvbSAnYXdzLWNkay1saWIvYXdzLWlhbSc7XG5pbXBvcnQgeyBDb25zdHJ1Y3QgfSBmcm9tICdjb25zdHJ1Y3RzJztcblxuZXhwb3J0IGludGVyZmFjZSBCdWlsZFRyaWdnZXJQcm9wcyB7XG4gIGJ1aWxkUHJvamVjdDogY29kZWJ1aWxkLlByb2plY3Q7XG59XG5cbmV4cG9ydCBjbGFzcyBPYnNpZGlhbkJ1aWxkVHJpZ2dlciBleHRlbmRzIENvbnN0cnVjdCB7XG4gIHB1YmxpYyByZWFkb25seSB0cmlnZ2VyRnVuY3Rpb246IGxhbWJkYS5GdW5jdGlvbjtcblxuICBjb25zdHJ1Y3RvcihzY29wZTogQ29uc3RydWN0LCBpZDogc3RyaW5nLCBwcm9wczogQnVpbGRUcmlnZ2VyUHJvcHMpIHtcbiAgICBzdXBlcihzY29wZSwgaWQpO1xuXG4gICAgLy8gQ3JlYXRlIExhbWJkYSBmdW5jdGlvbiB0byB0cmlnZ2VyIGJ1aWxkc1xuICAgIHRoaXMudHJpZ2dlckZ1bmN0aW9uID0gbmV3IGxhbWJkYS5GdW5jdGlvbih0aGlzLCAnQnVpbGRUcmlnZ2VyRnVuY3Rpb24nLCB7XG4gICAgICBydW50aW1lOiBsYW1iZGEuUnVudGltZS5OT0RFSlNfMThfWCxcbiAgICAgIGhhbmRsZXI6ICdpbmRleC5oYW5kbGVyJyxcbiAgICAgIGNvZGU6IGxhbWJkYS5Db2RlLmZyb21JbmxpbmUoYFxuICAgICAgICBjb25zdCBBV1MgPSByZXF1aXJlKCdhd3Mtc2RrJyk7XG4gICAgICAgIGNvbnN0IGNvZGVidWlsZCA9IG5ldyBBV1MuQ29kZUJ1aWxkKCk7XG4gICAgICAgIFxuICAgICAgICBleHBvcnRzLmhhbmRsZXIgPSBhc3luYyAoZXZlbnQpID0+IHtcbiAgICAgICAgICBjb25zdCBwYXJhbXMgPSB7XG4gICAgICAgICAgICBwcm9qZWN0TmFtZTogJyR7cHJvcHMuYnVpbGRQcm9qZWN0LnByb2plY3ROYW1lfScsXG4gICAgICAgICAgICBzb3VyY2VWZXJzaW9uOiBldmVudC5icmFuY2ggfHwgJ21haW4nXG4gICAgICAgICAgfTtcbiAgICAgICAgICBcbiAgICAgICAgICB0cnkge1xuICAgICAgICAgICAgY29uc3QgcmVzdWx0ID0gYXdhaXQgY29kZWJ1aWxkLnN0YXJ0QnVpbGQocGFyYW1zKS5wcm9taXNlKCk7XG4gICAgICAgICAgICByZXR1cm4ge1xuICAgICAgICAgICAgICBzdGF0dXNDb2RlOiAyMDAsXG4gICAgICAgICAgICAgIGJvZHk6IEpTT04uc3RyaW5naWZ5KHtcbiAgICAgICAgICAgICAgICBtZXNzYWdlOiAnQnVpbGQgc3RhcnRlZCBzdWNjZXNzZnVsbHknLFxuICAgICAgICAgICAgICAgIGJ1aWxkSWQ6IHJlc3VsdC5idWlsZC5pZCxcbiAgICAgICAgICAgICAgICBidWlsZEFybjogcmVzdWx0LmJ1aWxkLmFyblxuICAgICAgICAgICAgICB9KVxuICAgICAgICAgICAgfTtcbiAgICAgICAgICB9IGNhdGNoIChlcnJvcikge1xuICAgICAgICAgICAgcmV0dXJuIHtcbiAgICAgICAgICAgICAgc3RhdHVzQ29kZTogNTAwLFxuICAgICAgICAgICAgICBib2R5OiBKU09OLnN0cmluZ2lmeSh7XG4gICAgICAgICAgICAgICAgbWVzc2FnZTogJ0ZhaWxlZCB0byBzdGFydCBidWlsZCcsXG4gICAgICAgICAgICAgICAgZXJyb3I6IGVycm9yLm1lc3NhZ2VcbiAgICAgICAgICAgICAgfSlcbiAgICAgICAgICAgIH07XG4gICAgICAgICAgfVxuICAgICAgICB9O1xuICAgICAgYCksXG4gICAgICB0aW1lb3V0OiBjZGsuRHVyYXRpb24uc2Vjb25kcygzMCksXG4gICAgICBmdW5jdGlvbk5hbWU6ICdvYnNpZGlhbi1idWlsZC10cmlnZ2VyJyxcbiAgICB9KTtcblxuICAgIC8vIEdyYW50IHBlcm1pc3Npb25zIHRvIHRyaWdnZXIgYnVpbGRzXG4gICAgdGhpcy50cmlnZ2VyRnVuY3Rpb24uYWRkVG9Sb2xlUG9saWN5KG5ldyBpYW0uUG9saWN5U3RhdGVtZW50KHtcbiAgICAgIGVmZmVjdDogaWFtLkVmZmVjdC5BTExPVyxcbiAgICAgIGFjdGlvbnM6IFtcbiAgICAgICAgJ2NvZGVidWlsZDpTdGFydEJ1aWxkJyxcbiAgICAgICAgJ2NvZGVidWlsZDpCYXRjaEdldEJ1aWxkcycsXG4gICAgICBdLFxuICAgICAgcmVzb3VyY2VzOiBbcHJvcHMuYnVpbGRQcm9qZWN0LnByb2plY3RBcm5dLFxuICAgIH0pKTtcblxuICAgIC8vIE91dHB1dCB0aGUgZnVuY3Rpb24gbmFtZSBmb3IgZWFzeSBpbnZvY2F0aW9uXG4gICAgbmV3IGNkay5DZm5PdXRwdXQodGhpcywgJ0J1aWxkVHJpZ2dlckZ1bmN0aW9uTmFtZScsIHtcbiAgICAgIHZhbHVlOiB0aGlzLnRyaWdnZXJGdW5jdGlvbi5mdW5jdGlvbk5hbWUsXG4gICAgICBkZXNjcmlwdGlvbjogJ0xhbWJkYSBmdW5jdGlvbiB0byB0cmlnZ2VyIGJ1aWxkcycsXG4gICAgfSk7XG5cbiAgICAvLyBPdXRwdXQgY29tbWFuZCB0byB0cmlnZ2VyIGJ1aWxkXG4gICAgbmV3IGNkay5DZm5PdXRwdXQodGhpcywgJ1RyaWdnZXJCdWlsZENvbW1hbmQnLCB7XG4gICAgICB2YWx1ZTogYGF3cyBsYW1iZGEgaW52b2tlIC0tZnVuY3Rpb24tbmFtZSAke3RoaXMudHJpZ2dlckZ1bmN0aW9uLmZ1bmN0aW9uTmFtZX0gLS1wYXlsb2FkICd7XCJicmFuY2hcIjpcIm1haW5cIn0nIHJlc3BvbnNlLmpzb25gLFxuICAgICAgZGVzY3JpcHRpb246ICdDb21tYW5kIHRvIHRyaWdnZXIgYSBidWlsZCcsXG4gICAgfSk7XG4gIH1cbn0iXX0=