// Code generated by private/model/cli/gen-api/main.go. DO NOT EDIT.

// Package quicksightiface provides an interface to enable mocking the Amazon QuickSight service client
// for testing your code.
//
// It is important to note that this interface will have breaking changes
// when the service model is updated and adds new API operations, paginators,
// and waiters.
package quicksightiface

import (
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
)

// QuickSightAPI provides an interface to enable mocking the
// quicksight.QuickSight service client's API operation,
// paginators, and waiters. This make unit testing your code that calls out
// to the SDK's service client's calls easier.
//
// The best way to use this interface is so the SDK's service client's calls
// can be stubbed out for unit testing your code with the SDK without needing
// to inject custom request handlers into the SDK's request pipeline.
//
//    // myFunc uses an SDK service client to make a request to
//    // Amazon QuickSight.
//    func myFunc(svc quicksightiface.QuickSightAPI) bool {
//        // Make svc.CreateGroup request
//    }
//
//    func main() {
//        cfg, err := external.LoadDefaultAWSConfig()
//        if err != nil {
//            panic("failed to load config, " + err.Error())
//        }
//
//        svc := quicksight.New(cfg)
//
//        myFunc(svc)
//    }
//
// In your _test.go file:
//
//    // Define a mock struct to be used in your unit tests of myFunc.
//    type mockQuickSightClient struct {
//        quicksightiface.QuickSightAPI
//    }
//    func (m *mockQuickSightClient) CreateGroup(input *quicksight.CreateGroupInput) (*quicksight.CreateGroupOutput, error) {
//        // mock response/functionality
//    }
//
//    func TestMyFunc(t *testing.T) {
//        // Setup Test
//        mockSvc := &mockQuickSightClient{}
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
type QuickSightAPI interface {
	CreateGroupRequest(*quicksight.CreateGroupInput) quicksight.CreateGroupRequest

	CreateGroupMembershipRequest(*quicksight.CreateGroupMembershipInput) quicksight.CreateGroupMembershipRequest

	DeleteGroupRequest(*quicksight.DeleteGroupInput) quicksight.DeleteGroupRequest

	DeleteGroupMembershipRequest(*quicksight.DeleteGroupMembershipInput) quicksight.DeleteGroupMembershipRequest

	DeleteUserRequest(*quicksight.DeleteUserInput) quicksight.DeleteUserRequest

	DescribeGroupRequest(*quicksight.DescribeGroupInput) quicksight.DescribeGroupRequest

	DescribeUserRequest(*quicksight.DescribeUserInput) quicksight.DescribeUserRequest

	GetDashboardEmbedUrlRequest(*quicksight.GetDashboardEmbedUrlInput) quicksight.GetDashboardEmbedUrlRequest

	ListGroupMembershipsRequest(*quicksight.ListGroupMembershipsInput) quicksight.ListGroupMembershipsRequest

	ListGroupsRequest(*quicksight.ListGroupsInput) quicksight.ListGroupsRequest

	ListUserGroupsRequest(*quicksight.ListUserGroupsInput) quicksight.ListUserGroupsRequest

	ListUsersRequest(*quicksight.ListUsersInput) quicksight.ListUsersRequest

	RegisterUserRequest(*quicksight.RegisterUserInput) quicksight.RegisterUserRequest

	UpdateGroupRequest(*quicksight.UpdateGroupInput) quicksight.UpdateGroupRequest

	UpdateUserRequest(*quicksight.UpdateUserInput) quicksight.UpdateUserRequest
}

var _ QuickSightAPI = (*quicksight.QuickSight)(nil)
