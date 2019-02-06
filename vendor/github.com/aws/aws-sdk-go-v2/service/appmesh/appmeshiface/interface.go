// Code generated by private/model/cli/gen-api/main.go. DO NOT EDIT.

// Package appmeshiface provides an interface to enable mocking the AWS App Mesh service client
// for testing your code.
//
// It is important to note that this interface will have breaking changes
// when the service model is updated and adds new API operations, paginators,
// and waiters.
package appmeshiface

import (
	"github.com/aws/aws-sdk-go-v2/service/appmesh"
)

// AppMeshAPI provides an interface to enable mocking the
// appmesh.AppMesh service client's API operation,
// paginators, and waiters. This make unit testing your code that calls out
// to the SDK's service client's calls easier.
//
// The best way to use this interface is so the SDK's service client's calls
// can be stubbed out for unit testing your code with the SDK without needing
// to inject custom request handlers into the SDK's request pipeline.
//
//    // myFunc uses an SDK service client to make a request to
//    // AWS App Mesh.
//    func myFunc(svc appmeshiface.AppMeshAPI) bool {
//        // Make svc.CreateMesh request
//    }
//
//    func main() {
//        cfg, err := external.LoadDefaultAWSConfig()
//        if err != nil {
//            panic("failed to load config, " + err.Error())
//        }
//
//        svc := appmesh.New(cfg)
//
//        myFunc(svc)
//    }
//
// In your _test.go file:
//
//    // Define a mock struct to be used in your unit tests of myFunc.
//    type mockAppMeshClient struct {
//        appmeshiface.AppMeshAPI
//    }
//    func (m *mockAppMeshClient) CreateMesh(input *appmesh.CreateMeshInput) (*appmesh.CreateMeshOutput, error) {
//        // mock response/functionality
//    }
//
//    func TestMyFunc(t *testing.T) {
//        // Setup Test
//        mockSvc := &mockAppMeshClient{}
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
type AppMeshAPI interface {
	CreateMeshRequest(*appmesh.CreateMeshInput) appmesh.CreateMeshRequest

	CreateRouteRequest(*appmesh.CreateRouteInput) appmesh.CreateRouteRequest

	CreateVirtualNodeRequest(*appmesh.CreateVirtualNodeInput) appmesh.CreateVirtualNodeRequest

	CreateVirtualRouterRequest(*appmesh.CreateVirtualRouterInput) appmesh.CreateVirtualRouterRequest

	DeleteMeshRequest(*appmesh.DeleteMeshInput) appmesh.DeleteMeshRequest

	DeleteRouteRequest(*appmesh.DeleteRouteInput) appmesh.DeleteRouteRequest

	DeleteVirtualNodeRequest(*appmesh.DeleteVirtualNodeInput) appmesh.DeleteVirtualNodeRequest

	DeleteVirtualRouterRequest(*appmesh.DeleteVirtualRouterInput) appmesh.DeleteVirtualRouterRequest

	DescribeMeshRequest(*appmesh.DescribeMeshInput) appmesh.DescribeMeshRequest

	DescribeRouteRequest(*appmesh.DescribeRouteInput) appmesh.DescribeRouteRequest

	DescribeVirtualNodeRequest(*appmesh.DescribeVirtualNodeInput) appmesh.DescribeVirtualNodeRequest

	DescribeVirtualRouterRequest(*appmesh.DescribeVirtualRouterInput) appmesh.DescribeVirtualRouterRequest

	ListMeshesRequest(*appmesh.ListMeshesInput) appmesh.ListMeshesRequest

	ListRoutesRequest(*appmesh.ListRoutesInput) appmesh.ListRoutesRequest

	ListVirtualNodesRequest(*appmesh.ListVirtualNodesInput) appmesh.ListVirtualNodesRequest

	ListVirtualRoutersRequest(*appmesh.ListVirtualRoutersInput) appmesh.ListVirtualRoutersRequest

	UpdateRouteRequest(*appmesh.UpdateRouteInput) appmesh.UpdateRouteRequest

	UpdateVirtualNodeRequest(*appmesh.UpdateVirtualNodeInput) appmesh.UpdateVirtualNodeRequest

	UpdateVirtualRouterRequest(*appmesh.UpdateVirtualRouterInput) appmesh.UpdateVirtualRouterRequest
}

var _ AppMeshAPI = (*appmesh.AppMesh)(nil)
