// Code generated by private/model/cli/gen-api/main.go. DO NOT EDIT.

// Package lambdaiface provides an interface to enable mocking the AWS Lambda service client
// for testing your code.
//
// It is important to note that this interface will have breaking changes
// when the service model is updated and adds new API operations, paginators,
// and waiters.
package lambdaiface

import (
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

// LambdaAPI provides an interface to enable mocking the
// lambda.Lambda service client's API operation,
// paginators, and waiters. This make unit testing your code that calls out
// to the SDK's service client's calls easier.
//
// The best way to use this interface is so the SDK's service client's calls
// can be stubbed out for unit testing your code with the SDK without needing
// to inject custom request handlers into the SDK's request pipeline.
//
//    // myFunc uses an SDK service client to make a request to
//    // AWS Lambda.
//    func myFunc(svc lambdaiface.LambdaAPI) bool {
//        // Make svc.AddPermission request
//    }
//
//    func main() {
//        cfg, err := external.LoadDefaultAWSConfig()
//        if err != nil {
//            panic("failed to load config, " + err.Error())
//        }
//
//        svc := lambda.New(cfg)
//
//        myFunc(svc)
//    }
//
// In your _test.go file:
//
//    // Define a mock struct to be used in your unit tests of myFunc.
//    type mockLambdaClient struct {
//        lambdaiface.LambdaAPI
//    }
//    func (m *mockLambdaClient) AddPermission(input *lambda.AddPermissionInput) (*lambda.AddPermissionOutput, error) {
//        // mock response/functionality
//    }
//
//    func TestMyFunc(t *testing.T) {
//        // Setup Test
//        mockSvc := &mockLambdaClient{}
//
//        myfunc(mockSvc)
//
//        // Verify myFunc's functionality
//    }
//
// It is important to note that this interface will have breaking changes
// when the service model is updated and adds new API operations, paginators,
// and waiters. Its suggested to use the pattern above for testing, or using
// tooling to generate mocks to satisfy the interfaces.
type LambdaAPI interface {
	AddPermissionRequest(*lambda.AddPermissionInput) lambda.AddPermissionRequest

	CreateAliasRequest(*lambda.CreateAliasInput) lambda.CreateAliasRequest

	CreateEventSourceMappingRequest(*lambda.CreateEventSourceMappingInput) lambda.CreateEventSourceMappingRequest

	CreateFunctionRequest(*lambda.CreateFunctionInput) lambda.CreateFunctionRequest

	DeleteAliasRequest(*lambda.DeleteAliasInput) lambda.DeleteAliasRequest

	DeleteEventSourceMappingRequest(*lambda.DeleteEventSourceMappingInput) lambda.DeleteEventSourceMappingRequest

	DeleteFunctionRequest(*lambda.DeleteFunctionInput) lambda.DeleteFunctionRequest

	DeleteFunctionConcurrencyRequest(*lambda.DeleteFunctionConcurrencyInput) lambda.DeleteFunctionConcurrencyRequest

	GetAccountSettingsRequest(*lambda.GetAccountSettingsInput) lambda.GetAccountSettingsRequest

	GetAliasRequest(*lambda.GetAliasInput) lambda.GetAliasRequest

	GetEventSourceMappingRequest(*lambda.GetEventSourceMappingInput) lambda.GetEventSourceMappingRequest

	GetFunctionRequest(*lambda.GetFunctionInput) lambda.GetFunctionRequest

	GetFunctionConfigurationRequest(*lambda.GetFunctionConfigurationInput) lambda.GetFunctionConfigurationRequest

	GetPolicyRequest(*lambda.GetPolicyInput) lambda.GetPolicyRequest

	InvokeRequest(*lambda.InvokeInput) lambda.InvokeRequest

	InvokeAsyncRequest(*lambda.InvokeAsyncInput) lambda.InvokeAsyncRequest

	ListAliasesRequest(*lambda.ListAliasesInput) lambda.ListAliasesRequest

	ListEventSourceMappingsRequest(*lambda.ListEventSourceMappingsInput) lambda.ListEventSourceMappingsRequest

	ListFunctionsRequest(*lambda.ListFunctionsInput) lambda.ListFunctionsRequest

	ListTagsRequest(*lambda.ListTagsInput) lambda.ListTagsRequest

	ListVersionsByFunctionRequest(*lambda.ListVersionsByFunctionInput) lambda.ListVersionsByFunctionRequest

	PublishVersionRequest(*lambda.PublishVersionInput) lambda.PublishVersionRequest

	PutFunctionConcurrencyRequest(*lambda.PutFunctionConcurrencyInput) lambda.PutFunctionConcurrencyRequest

	RemovePermissionRequest(*lambda.RemovePermissionInput) lambda.RemovePermissionRequest

	TagResourceRequest(*lambda.TagResourceInput) lambda.TagResourceRequest

	UntagResourceRequest(*lambda.UntagResourceInput) lambda.UntagResourceRequest

	UpdateAliasRequest(*lambda.UpdateAliasInput) lambda.UpdateAliasRequest

	UpdateEventSourceMappingRequest(*lambda.UpdateEventSourceMappingInput) lambda.UpdateEventSourceMappingRequest

	UpdateFunctionCodeRequest(*lambda.UpdateFunctionCodeInput) lambda.UpdateFunctionCodeRequest

	UpdateFunctionConfigurationRequest(*lambda.UpdateFunctionConfigurationInput) lambda.UpdateFunctionConfigurationRequest
}

var _ LambdaAPI = (*lambda.Lambda)(nil)