package iac

import (
	"fmt"
	"strings"
	"time"
)

// CloudFormationGenerator generates CloudFormation templates for AWS optimizations
type CloudFormationGenerator struct {
	region string
}

// NewCloudFormationGenerator creates a new CloudFormation generator
func NewCloudFormationGenerator(region string) *CloudFormationGenerator {
	return &CloudFormationGenerator{
		region: region,
	}
}

// GenerateEC2Optimization generates CloudFormation for EC2 optimization
func (cfg *CloudFormationGenerator) GenerateEC2Optimization(recommendation *OptimizationRecommendation) *IaCScript {
	var cfTemplate strings.Builder
	
	// Template header
	cfTemplate.WriteString(fmt.Sprintf(`AWSTemplateFormatVersion: '2010-09-09'
Description: 'Yukti FinOps - EC2 Optimization Template'
# Generated: %s
# Action: %s
# Estimated Savings: $%.2f/month

Parameters:
  InstanceId:
    Type: String
    Default: '%s'
    Description: 'Target EC2 instance ID'

`, time.Now().Format("2006-01-02 15:04:05"), recommendation.Action, recommendation.EstimatedSavings*30, recommendation.ResourceID))

	switch recommendation.Action {
	case "downsize":
		cfTemplate.WriteString(cfg.generateDownsizeTemplate(recommendation))
	case "terminate":
		cfTemplate.WriteString(cfg.generateTerminateTemplate(recommendation))
	case "schedule":
		cfTemplate.WriteString(cfg.generateScheduleTemplate(recommendation))
	case "spot_conversion":
		cfTemplate.WriteString(cfg.generateSpotTemplate(recommendation))
	}

	return &IaCScript{
		ID:               fmt.Sprintf("cf-%s-%d", recommendation.ResourceID, time.Now().Unix()),
		Type:             "cloudformation",
		Provider:         "aws",
		Action:           recommendation.Action,
		ResourceID:       recommendation.ResourceID,
		Code:             cfTemplate.String(),
		RollbackCode:     cfg.generateRollbackTemplate(recommendation),
		EstimatedSavings: recommendation.EstimatedSavings,
		GeneratedAt:      time.Now(),
		Instructions:     cfg.generateInstructions(recommendation),
	}
}

// generateDownsizeTemplate generates CloudFormation for instance downsizing
func (cfg *CloudFormationGenerator) generateDownsizeTemplate(rec *OptimizationRecommendation) string {
	return fmt.Sprintf(`Resources:
  InstanceModification:
    Type: AWS::EC2::Instance
    Properties:
      InstanceType: %s
      ImageId: !Ref OriginalAMI
      KeyName: !Ref OriginalKeyName
      SecurityGroupIds: !Ref OriginalSecurityGroups
      SubnetId: !Ref OriginalSubnet
      Tags:
        - Key: 'OptimizedBy'
          Value: 'Yukti-FinOps'
        - Key: 'OriginalInstanceType'
          Value: !Ref OriginalInstanceType
        - Key: 'EstimatedMonthlySavings'
          Value: '%.2f'

  # Lambda function to handle instance replacement
  InstanceReplacer:
    Type: AWS::Lambda::Function
    Properties:
      FunctionName: !Sub 'instance-replacer-${InstanceId}'
      Runtime: python3.9
      Handler: index.handler
      Code:
        ZipFile: |
          import boto3
          import json
          
          def handler(event, context):
              ec2 = boto3.client('ec2')
              instance_id = event['InstanceId']
              new_instance_type = event['NewInstanceType']
              
              # Stop instance
              ec2.stop_instances(InstanceIds=[instance_id])
              
              # Wait for stopped state
              waiter = ec2.get_waiter('instance_stopped')
              waiter.wait(InstanceIds=[instance_id])
              
              # Modify instance type
              ec2.modify_instance_attribute(
                  InstanceId=instance_id,
                  InstanceType={'Value': new_instance_type}
              )
              
              return {'statusCode': 200, 'body': 'Instance optimized'}
      Role: !GetAtt LambdaExecutionRole.Arn

  LambdaExecutionRole:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Principal:
              Service: lambda.amazonaws.com
            Action: sts:AssumeRole
      ManagedPolicyArns:
        - arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
      Policies:
        - PolicyName: EC2ModifyPolicy
          PolicyDocument:
            Version: '2012-10-17'
            Statement:
              - Effect: Allow
                Action:
                  - ec2:DescribeInstances
                  - ec2:StopInstances
                  - ec2:StartInstances
                  - ec2:ModifyInstanceAttribute
                Resource: '*'

Outputs:
  OptimizedInstanceType:
    Description: 'New optimized instance type'
    Value: %s
  EstimatedMonthlySavings:
    Description: 'Estimated monthly cost savings'
    Value: !Sub '%.2f USD'
`, rec.RecommendedInstanceType, rec.EstimatedSavings*30, rec.RecommendedInstanceType, rec.EstimatedSavings*30)
}

// generateTerminateTemplate generates CloudFormation for instance termination
func (cfg *CloudFormationGenerator) generateTerminateTemplate(rec *OptimizationRecommendation) string {
	return fmt.Sprintf(`Resources:
  # Create AMI backup before termination
  InstanceBackup:
    Type: AWS::EC2::Image
    Properties:
      Name: !Sub 'backup-${InstanceId}-${AWS::StackName}'
      InstanceId: !Ref InstanceId
      Description: 'Pre-termination backup created by Yukti FinOps'
      Tags:
        - Key: 'Purpose'
          Value: 'Pre-termination backup'
        - Key: 'CreatedBy'
          Value: 'Yukti-FinOps'
        - Key: 'OriginalInstance'
          Value: !Ref InstanceId

  # Lambda function for safe termination
  TerminationFunction:
    Type: AWS::Lambda::Function
    Properties:
      FunctionName: !Sub 'safe-terminator-${InstanceId}'
      Runtime: python3.9
      Handler: index.handler
      Code:
        ZipFile: |
          import boto3
          import json
          
          def handler(event, context):
              ec2 = boto3.client('ec2')
              instance_id = event['InstanceId']
              
              # Verify backup AMI exists
              images = ec2.describe_images(
                  Owners=['self'],
                  Filters=[
                      {'Name': 'tag:OriginalInstance', 'Values': [instance_id]}
                  ]
              )
              
              if not images['Images']:
                  return {'statusCode': 400, 'body': 'No backup AMI found'}
              
              # Terminate instance (uncomment when ready)
              # ec2.terminate_instances(InstanceIds=[instance_id])
              
              return {'statusCode': 200, 'body': 'Ready for termination'}
      Role: !GetAtt LambdaExecutionRole.Arn

  LambdaExecutionRole:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Principal:
              Service: lambda.amazonaws.com
            Action: sts:AssumeRole
      ManagedPolicyArns:
        - arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
      Policies:
        - PolicyName: EC2TerminatePolicy
          PolicyDocument:
            Version: '2012-10-17'
            Statement:
              - Effect: Allow
                Action:
                  - ec2:DescribeInstances
                  - ec2:DescribeImages
                  - ec2:TerminateInstances
                  - ec2:CreateImage
                Resource: '*'

Outputs:
  BackupAMI:
    Description: 'Backup AMI ID'
    Value: !Ref InstanceBackup
  EstimatedMonthlySavings:
    Description: 'Estimated monthly cost savings'
    Value: '%.2f USD'
`, rec.EstimatedSavings*30)
}

// generateScheduleTemplate generates CloudFormation for instance scheduling
func (cfg *CloudFormationGenerator) generateScheduleTemplate(rec *OptimizationRecommendation) string {
	return `Resources:
  # Lambda function for instance scheduling
  SchedulerFunction:
    Type: AWS::Lambda::Function
    Properties:
      FunctionName: !Sub 'ec2-scheduler-${InstanceId}'
      Runtime: python3.9
      Handler: index.handler
      Environment:
        Variables:
          INSTANCE_ID: !Ref InstanceId
      Code:
        ZipFile: |
          import boto3
          import os
          
          def handler(event, context):
              ec2 = boto3.client('ec2')
              instance_id = os.environ['INSTANCE_ID']
              action = event.get('action', 'stop')
              
              if action == 'start':
                  ec2.start_instances(InstanceIds=[instance_id])
              elif action == 'stop':
                  ec2.stop_instances(InstanceIds=[instance_id])
              
              return {'statusCode': 200, 'body': f'Instance {action}ped'}
      Role: !GetAtt SchedulerRole.Arn

  SchedulerRole:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Principal:
              Service: lambda.amazonaws.com
            Action: sts:AssumeRole
      ManagedPolicyArns:
        - arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
      Policies:
        - PolicyName: EC2SchedulePolicy
          PolicyDocument:
            Version: '2012-10-17'
            Statement:
              - Effect: Allow
                Action:
                  - ec2:StartInstances
                  - ec2:StopInstances
                  - ec2:DescribeInstances
                Resource: '*'

  # EventBridge rules for scheduling
  StopSchedule:
    Type: AWS::Events::Rule
    Properties:
      Name: !Sub 'stop-${InstanceId}'
      Description: 'Stop instance at 6 PM daily'
      ScheduleExpression: 'cron(0 18 * * ? *)'
      State: ENABLED
      Targets:
        - Arn: !GetAtt SchedulerFunction.Arn
          Id: 'StopTarget'
          Input: '{"action": "stop"}'

  StartSchedule:
    Type: AWS::Events::Rule
    Properties:
      Name: !Sub 'start-${InstanceId}'
      Description: 'Start instance at 8 AM daily'
      ScheduleExpression: 'cron(0 8 * * ? *)'
      State: ENABLED
      Targets:
        - Arn: !GetAtt SchedulerFunction.Arn
          Id: 'StartTarget'
          Input: '{"action": "start"}'

  # Permissions for EventBridge to invoke Lambda
  StopSchedulePermission:
    Type: AWS::Lambda::Permission
    Properties:
      FunctionName: !Ref SchedulerFunction
      Action: lambda:InvokeFunction
      Principal: events.amazonaws.com
      SourceArn: !GetAtt StopSchedule.Arn

  StartSchedulePermission:
    Type: AWS::Lambda::Permission
    Properties:
      FunctionName: !Ref SchedulerFunction
      Action: lambda:InvokeFunction
      Principal: events.amazonaws.com
      SourceArn: !GetAtt StartSchedule.Arn

Outputs:
  SchedulerFunction:
    Description: 'Lambda function for scheduling'
    Value: !Ref SchedulerFunction
  EstimatedMonthlySavings:
    Description: 'Estimated monthly savings (58% uptime reduction)'
    Value: '58% cost reduction'`
}

// generateSpotTemplate generates CloudFormation for Spot conversion
func (cfg *CloudFormationGenerator) generateSpotTemplate(rec *OptimizationRecommendation) string {
	return fmt.Sprintf(`Resources:
  # Launch template for Spot instances
  SpotLaunchTemplate:
    Type: AWS::EC2::LaunchTemplate
    Properties:
      LaunchTemplateName: !Sub 'spot-template-${InstanceId}'
      LaunchTemplateData:
        ImageId: !Ref OriginalAMI
        InstanceType: !Ref OriginalInstanceType
        KeyName: !Ref OriginalKeyName
        SecurityGroupIds: !Ref OriginalSecurityGroups
        InstanceMarketOptions:
          MarketType: spot
          SpotOptions:
            MaxPrice: '%.4f'
        TagSpecifications:
          - ResourceType: instance
            Tags:
              - Key: 'SpotInstance'
                Value: 'true'
              - Key: 'ConvertedBy'
                Value: 'Yukti-FinOps'
              - Key: 'OriginalInstance'
                Value: !Ref InstanceId

  # Auto Scaling Group for Spot instance
  SpotAutoScalingGroup:
    Type: AWS::AutoScaling::AutoScalingGroup
    Properties:
      AutoScalingGroupName: !Sub 'spot-asg-${InstanceId}'
      VPCZoneIdentifier: 
        - !Ref OriginalSubnet
      MinSize: 1
      MaxSize: 1
      DesiredCapacity: 1
      LaunchTemplate:
        LaunchTemplateId: !Ref SpotLaunchTemplate
        Version: !GetAtt SpotLaunchTemplate.LatestVersionNumber

Outputs:
  SpotLaunchTemplate:
    Description: 'Launch template for Spot instances'
    Value: !Ref SpotLaunchTemplate
  EstimatedMonthlySavings:
    Description: 'Estimated monthly savings (70%% cost reduction)'
    Value: '%.2f USD'
`, rec.EstimatedSavings*0.7, rec.EstimatedSavings*30)
}

// generateRollbackTemplate generates rollback CloudFormation template
func (cfg *CloudFormationGenerator) generateRollbackTemplate(rec *OptimizationRecommendation) string {
	return fmt.Sprintf(`AWSTemplateFormatVersion: '2010-09-09'
Description: 'ROLLBACK TEMPLATE - Yukti FinOps'
# Use this template to rollback optimization changes
# Generated for: %s

Parameters:
  InstanceId:
    Type: String
    Default: '%s'
    Description: 'Instance ID to rollback'

Resources:
  RollbackFunction:
    Type: AWS::Lambda::Function
    Properties:
      FunctionName: !Sub 'rollback-${InstanceId}'
      Runtime: python3.9
      Handler: index.handler
      Code:
        ZipFile: |
          import boto3
          import json
          
          def handler(event, context):
              # Implement rollback logic based on the optimization type
              # This is a template - customize based on your specific needs
              return {'statusCode': 200, 'body': 'Rollback completed'}
      Role: !GetAtt RollbackRole.Arn

  RollbackRole:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Principal:
              Service: lambda.amazonaws.com
            Action: sts:AssumeRole
      ManagedPolicyArns:
        - arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole

Outputs:
  RollbackFunction:
    Description: 'Function to rollback changes'
    Value: !Ref RollbackFunction
`, rec.ResourceID, rec.ResourceID)
}

// generateInstructions generates CloudFormation deployment instructions
func (cfg *CloudFormationGenerator) generateInstructions(rec *OptimizationRecommendation) []string {
	return []string{
		"1. Review the CloudFormation template carefully",
		"2. Ensure you have appropriate AWS permissions",
		"3. Deploy using AWS CLI: aws cloudformation create-stack",
		"4. Or use AWS Console CloudFormation service",
		"5. Monitor the stack creation progress",
		"6. Verify the optimization was applied correctly",
		"7. Keep the rollback template for emergency restoration",
		fmt.Sprintf("8. Expected savings: $%.2f per month", rec.EstimatedSavings*30),
	}
}