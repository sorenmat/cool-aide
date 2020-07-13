package rds

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/pkg/errors"
	"github.com/sorenmat/k8s-rds/crd"
	"github.com/sorenmat/k8s-rds/provider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type RDS struct {
	EC2             *ec2.Client
	Subnets         []string
	SecurityGroups  []string
	VpcId           string
	ServiceProvider provider.ServiceProvider
}

func New(db *crd.Database, kc *kubernetes.Clientset) (*RDS, error) {
	ec2client, err := ec2client(kc)
	if err != nil {
		log.Fatal("Unable to create a client for EC2 ", err)
	}

	nodeInfo, err := describeNodeEC2Instance(kc, ec2client)
	if err != nil {
		log.Println(err)
		return nil, errors.Wrap(err, "Unable AWS metadata")
	}

	vpcId := *nodeInfo.Reservations[0].Instances[0].VpcId

	log.Println("Trying to get subnets")
	subnets := db.Spec.Subnets
	if db.Spec.DBSubnetGroupName == "" || len(subnets) < 2 {
		subnets, err = getSubnets(nodeInfo, ec2client, db.Spec.PubliclyAccessible)
		if err != nil {
			return nil, fmt.Errorf("unable to get subnets from instance: %v", err)

		}
	} else {
		log.Println("Got the following Subnets from spec")
		for _, v := range subnets {
			log.Printf(v + " ")
		}
	}
	sgs := db.Spec.SecurityGroups
	if len(sgs) == 0 {
		log.Println("Trying to get security groups")
		sgs, err = getSGS(kc, ec2client)
		if err != nil {
			return nil, fmt.Errorf("unable to get security groups from instance: %v", err)

		}
	} else {
		log.Println("Got Security Groups from spec")
	}

	r := RDS{EC2: ec2client, Subnets: subnets, SecurityGroups: sgs, VpcId: vpcId}
	return &r, nil
}

// CreateDatabase creates a database from the CRD database object, is also ensures that the correct
// subnets are created for the database so we can access it
func (r *RDS) CreateDatabase(db *crd.Database) (string, error) {
	if db.Spec.DBSnapshotIdentifier == "" {
		return r.CreateDatabaseInstance(db)
	} else {
		return r.RestoreDatabaseFromSnapshot(db)
	}
}

func (r *RDS) getSubnetGroupName(db *crd.Database) (string, error) {
	var err error
	dbSubnetGroupName := db.Spec.DBSubnetGroupName
	if dbSubnetGroupName == "" {
		// Ensure that the subnets for the DB is create or updated
		log.Println("Trying to find the correct subnets")
		dbSubnetGroupName, err = r.ensureSubnets(db)
		if err != nil {
			return "", err
		}
	}
	return dbSubnetGroupName, nil
}

// CreateDatabase creates a database from the CRD database object, is also ensures that the correct
// subnets are created for the database so we can access it
func (r *RDS) CreateDatabaseInstance(db *crd.Database) (string, error) {
	dbSubnetGroupName, err := r.getSubnetGroupName(db)

	log.Printf("getting secret: Name: %v Key: %v \n", db.Spec.Password.Name, db.Spec.Password.Key)
	pw, err := r.GetSecret(db.Namespace, db.Spec.Password.Name, db.Spec.Password.Key)
	if err != nil {
		return "", err
	}
	input := convertSpecToInstanceInput(db, dbSubnetGroupName, r.SecurityGroups, pw)

	// search for the instance
	log.Printf("Trying to find db instance %v\n", db.Spec.DBName)
	k := &rds.DescribeDBInstancesInput{DBInstanceIdentifier: input.DBInstanceIdentifier}
	res := r.rdsclient().DescribeDBInstancesRequest(k)
	_, err = res.Send(context.Background())
	if err != nil && err.Error() != rds.ErrCodeDBInstanceNotFoundFault {
		log.Printf("DB instance %v not found trying to create it\n", db.Spec.DBName)
		// seems like we didn't find a database with this name, let's create on
		res := r.rdsclient().CreateDBInstanceRequest(input)
		_, err = res.Send(context.Background())
		if err != nil {
			return "", errors.Wrap(err, "CreateDBInstance")
		}
	} else if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("wasn't able to describe the db instance with id %v", input.DBInstanceIdentifier))
	}
	log.Printf("Waiting for db instance %v to become available\n", *input.DBInstanceIdentifier)
	time.Sleep(5 * time.Second)
	err = r.rdsclient().WaitUntilDBInstanceAvailable(context.Background(), k)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("something went wrong in WaitUntilDBInstanceAvailable for db instance %v", input.DBInstanceIdentifier))
	}

	// Get the newly created database so we can get the endpoint
	dbHostname, err := getEndpoint(input.DBInstanceIdentifier, r.rdsclient())
	if err != nil {
		return "", err
	}
	return dbHostname, nil
}

// RestoreDatabaseFromSnapshot creates a database instance from a snapshot using the CRD database object, is also ensures that the correct
// subnets are created for the database so we can access it
func (r *RDS) RestoreDatabaseFromSnapshot(db *crd.Database) (string, error) {
	dbSubnetGroupName, err := r.getSubnetGroupName(db)

	log.Printf("getting secret: Name: %v Key: %v \n", db.Spec.Password.Name, db.Spec.Password.Key)
	pw, err := r.GetSecret(db.Namespace, db.Spec.Password.Name, db.Spec.Password.Key)
	if err != nil {
		return "", err
	}
	restoreSnapshotInput, modifyInstanceInput := convertSpecToRestoreSnapshotInput(db, dbSubnetGroupName, r.SecurityGroups, pw)

	// search for the instance
	log.Printf("Trying to find db instance %v\n", *restoreSnapshotInput.DBInstanceIdentifier)
	k := &rds.DescribeDBInstancesInput{DBInstanceIdentifier: restoreSnapshotInput.DBInstanceIdentifier}
	res := r.rdsclient().DescribeDBInstancesRequest(k)
	_, err = res.Send(context.Background())
	switch {
	case err == nil:
		return "", errors.New(fmt.Sprintf("DB instance %v already exists. Will not restore", *restoreSnapshotInput.DBInstanceIdentifier))
	case err.Error() == rds.ErrCodeDBInstanceNotFoundFault:
		log.Printf("DB instance %v not found trying to restore it\n", *restoreSnapshotInput.DBInstanceIdentifier)
		// seems like we didn't find a database with this name, let's create on
		res := r.rdsclient().RestoreDBInstanceFromDBSnapshotRequest(restoreSnapshotInput)
		_, err = res.Send(context.Background())
		if err != nil {
			return "", errors.Wrap(err, "RestoreDBInstanceFromDBSnapshot")
		}
	default:
		return "", errors.Wrap(err, fmt.Sprintf("wasn't able to describe the db instance with id %v", restoreSnapshotInput.DBInstanceIdentifier))
	}

	log.Printf("Waiting for db instance %v to become available\n", *restoreSnapshotInput.DBInstanceIdentifier)
	time.Sleep(5 * time.Second)
	err = r.rdsclient().WaitUntilDBInstanceAvailable(context.Background(), k)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("something went wrong in WaitUntilDBInstanceAvailable for db instance %v", *restoreSnapshotInput.DBInstanceIdentifier))
	}

	if modifyInstanceInput != nil {
		log.Printf("DB instance %v restored.\n", *restoreSnapshotInput.DBInstanceIdentifier)
	} else {
		// apply needed modifications
		log.Printf("DB instance %v restored. Applying some modifications\n", *restoreSnapshotInput.DBInstanceIdentifier)
		resModify := r.rdsclient().ModifyDBInstanceRequest(modifyInstanceInput)
		_, err = resModify.Send(context.Background())
		if err != nil {
			return "", errors.Wrap(err, "ModifyDBInstance")
		}
		log.Printf("Waiting for db instance %v to become available\n", *restoreSnapshotInput.DBInstanceIdentifier)
		time.Sleep(5 * time.Second)
		err = r.rdsclient().WaitUntilDBInstanceAvailable(context.Background(), k)
		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("something went wrong in WaitUntilDBInstanceAvailable for db instance %v", *restoreSnapshotInput.DBInstanceIdentifier))
		}
	}

	// Get the newly created database so we can get the endpoint
	dbHostname, err := getEndpoint(restoreSnapshotInput.DBInstanceIdentifier, r.rdsclient())
	if err != nil {
		return "", err
	}
	return dbHostname, nil
}

// ensureSubnets is ensuring that we have created or updated the subnet according to the data from the CRD object
func (r *RDS) ensureSubnets(db *crd.Database) (string, error) {
	if len(r.Subnets) == 0 {
		log.Println("Error: unable to continue due to lack of subnets, perhaps we couldn't lookup the subnets")
	}
	subnetDescription := "RDS Subnet Group for VPC: " + r.VpcId
	subnetName := "db-subnetgroup-" + r.VpcId

	svc := r.rdsclient()

	sf := &rds.DescribeDBSubnetGroupsInput{DBSubnetGroupName: aws.String(subnetName)}
	res := svc.DescribeDBSubnetGroupsRequest(sf)
	_, err := res.Send(context.Background())
	log.Println("Subnets:", r.Subnets)
	if err != nil {
		// assume we didn't find it..
		subnet := &rds.CreateDBSubnetGroupInput{
			DBSubnetGroupDescription: aws.String(subnetDescription),
			DBSubnetGroupName:        aws.String(subnetName),
			SubnetIds:                r.Subnets,
			Tags:                     []rds.Tag{{Key: aws.String("Warning"), Value: aws.String("Managed by k8s-rds.")}},
		}
		res := svc.CreateDBSubnetGroupRequest(subnet)
		_, err := res.Send(context.Background())
		if err != nil {
			return "", errors.Wrap(err, "CreateDBSubnetGroup")
		}
	} else {
		log.Printf("Moving on seems like %v exsits", subnetName)
	}
	return subnetName, nil
}

func getEndpoint(dbName *string, svc *rds.Client) (string, error) {
	k := &rds.DescribeDBInstancesInput{DBInstanceIdentifier: dbName}
	res := svc.DescribeDBInstancesRequest(k)
	instance, err := res.Send(context.Background())
	if err != nil || len(instance.DBInstances) == 0 {
		return "", fmt.Errorf("wasn't able to describe the db instance with id %v", dbName)
	}
	rdsdb := instance.DBInstances[0]

	dbHostname := *rdsdb.Endpoint.Address
	return dbHostname, nil
}

func (r *RDS) DeleteDatabase(db *crd.Database) error {
	if db.Spec.DeleteProtection {
		log.Printf("Trying to delete a %v in %v which is a deleted protected database", db.Name, db.Namespace)
		return nil
	}
	// delete the database instance
	svc := r.rdsclient()
	res := svc.DeleteDBInstanceRequest(&rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbidentifier(db)),
		SkipFinalSnapshot:    aws.Bool(true),
	})
	_, err := res.Send(context.Background())
	if err != nil {
		e := errors.Wrap(err, fmt.Sprintf("unable to delete database %v", db.Spec.DBName))
		log.Println(e)
		return e
	} else {
		log.Printf("Waiting for db instance %v to be deleted\n", db.Spec.DBName)
		time.Sleep(5 * time.Second)

		k := &rds.DescribeDBInstancesInput{DBInstanceIdentifier: aws.String(dbidentifier(db))}
		err = r.rdsclient().WaitUntilDBInstanceDeleted(context.Background(), k)
		if err != nil {
			log.Println(err)
			return err
		} else {
			log.Println("Deleted DB instance: ", db.Spec.DBName)
		}
	}

	// delete the subnet group attached to the instance
	subnetName := db.Name + "-subnet-" + db.Namespace
	dres := svc.DeleteDBSubnetGroupRequest(&rds.DeleteDBSubnetGroupInput{DBSubnetGroupName: aws.String(subnetName)})
	_, err = dres.Send(context.Background())
	if err != nil {
		e := errors.Wrap(err, fmt.Sprintf("unable to delete subnet %v", subnetName))
		log.Println(e)
		return e
	} else {
		log.Println("Deleted DBSubnet group: ", subnetName)
	}
	return nil
}

func (r *RDS) rdsclient() *rds.Client {
	return rds.New(r.EC2.Config)
}
func dbidentifier(v *crd.Database) string {
	return v.Name + "-" + v.Namespace
}

const (
	maxTagLengthAllowed = 255
	tagRegexp           = `^kube.*$`
)

func toTags(annotations, labels map[string]string) []rds.Tag {
	tags := []rds.Tag{}
	r := regexp.MustCompile(tagRegexp)

	for k, v := range annotations {
		if len(k) > maxTagLengthAllowed || len(v) > maxTagLengthAllowed ||
			r.Match([]byte(k)) {
			log.Printf("WARNING: Not Adding annotation KV to tags: %v %v", k, v)
			continue
		}

		tags = append(tags, rds.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	for k, v := range labels {
		if len(k) > maxTagLengthAllowed || len(v) > maxTagLengthAllowed {
			log.Printf("WARNING: Not Adding CRD labels KV to tags: %v %v", k, v)
			continue
		}

		tags = append(tags, rds.Tag{Key: aws.String(k), Value: aws.String(v)})
	}

	return tags
}

func gettags(db *crd.Database) []rds.Tag {
	tags := []rds.Tag{}
	if db.Spec.Tags == "" {
		return tags
	}
	for _, v := range strings.Split(db.Spec.Tags, ",") {
		kv := strings.Split(v, "=")

		tags = append(tags, rds.Tag{Key: aws.String(strings.TrimSpace(kv[0])), Value: aws.String(strings.TrimSpace(kv[1]))})
	}
	return tags
}

func convertSpecToInstanceInput(v *crd.Database, subnetName string, securityGroups []string, password string) *rds.CreateDBInstanceInput {
	tags := toTags(v.Annotations, v.Labels)
	tags = append(tags, gettags(v)...)

	input := &rds.CreateDBInstanceInput{
		DBName:                aws.String(v.Spec.DBName),
		AllocatedStorage:      aws.Int64(v.Spec.Size),
		DBInstanceClass:       aws.String(v.Spec.Class),
		DBInstanceIdentifier:  aws.String(dbidentifier(v)),
		VpcSecurityGroupIds:   securityGroups,
		Engine:                aws.String(v.Spec.Engine),
		MasterUserPassword:    aws.String(password),
		MasterUsername:        aws.String(v.Spec.Username),
		DBSubnetGroupName:     aws.String(subnetName),
		PubliclyAccessible:    aws.Bool(v.Spec.PubliclyAccessible),
		MultiAZ:               aws.Bool(v.Spec.MultiAZ),
		StorageEncrypted:      aws.Bool(v.Spec.StorageEncrypted),
		BackupRetentionPeriod: aws.Int64(v.Spec.BackupRetentionPeriod),
		DeletionProtection:    aws.Bool(v.Spec.DeleteProtection),
		Tags:                  tags,
	}
	if v.Spec.StorageType != "" {
		input.StorageType = aws.String(v.Spec.StorageType)
	}
	if v.Spec.Iops > 0 {
		input.Iops = aws.Int64(v.Spec.Iops)
	}
	return input
}

func convertSpecToRestoreSnapshotInput(v *crd.Database, dbSubnetGroupName string, securityGroups []string, password string) (*rds.RestoreDBInstanceFromDBSnapshotInput, *rds.ModifyDBInstanceInput) {
	restoreSnapshotInput := &rds.RestoreDBInstanceFromDBSnapshotInput{
		DBInstanceClass:      aws.String(v.Spec.Class),
		DBInstanceIdentifier: aws.String(v.Name + "-" + v.Namespace),
		DBSnapshotIdentifier: aws.String(v.Spec.DBSnapshotIdentifier),
		Engine:               aws.String(v.Spec.Engine),
		DBSubnetGroupName:    aws.String(dbSubnetGroupName),
		VpcSecurityGroupIds:  securityGroups,
		PubliclyAccessible:   aws.Bool(v.Spec.PubliclyAccessible),
		MultiAZ:              aws.Bool(v.Spec.MultiAZ),
	}

	switch v.Spec.Engine {
	case "mariadb":
		fallthrough
	case "mysql":
		fallthrough
	case "postgres":
		restoreSnapshotInput.DBName = aws.String("")
	default:
		restoreSnapshotInput.DBName = aws.String(v.Spec.DBName)
	}

	if v.Spec.StorageType != "" {
		restoreSnapshotInput.StorageType = aws.String(v.Spec.StorageType)
	}
	if v.Spec.Iops > 0 {
		restoreSnapshotInput.Iops = aws.Int64(v.Spec.Iops)
	}
	if v.Spec.DBParameterGroupName != "" {
		restoreSnapshotInput.DBParameterGroupName = aws.String(v.Spec.DBParameterGroupName)
	}

	modifyInstance := false
	modifyInstanceInput := &rds.ModifyDBInstanceInput{
		DBInstanceIdentifier: aws.String(v.Name + "-" + v.Namespace),
	}

	if password != "" {
		modifyInstanceInput.MasterUserPassword = aws.String(password)
		modifyInstance = true
	}
	if v.Spec.Size > 0 {
		modifyInstanceInput.AllocatedStorage = aws.Int64(v.Spec.Size)
		modifyInstance = true
	}

	if v.Spec.BackupRetentionPeriod != int64(0) {
		modifyInstanceInput.BackupRetentionPeriod = aws.Int64(v.Spec.BackupRetentionPeriod)
		modifyInstance = true
	}

	if modifyInstance {
		return restoreSnapshotInput, modifyInstanceInput
	}
	return restoreSnapshotInput, nil
}

// describeNodeEC2Instance returns the AWS Metadata for the first Node from the cluster
func describeNodeEC2Instance(kubectl *kubernetes.Clientset, svc *ec2.Client) (*ec2.DescribeInstancesResponse, error) {
	nodes, err := kubectl.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get nodes")
	}
	name := ""

	if len(nodes.Items) == 0 {
		return nil, fmt.Errorf("unable to find any nodes in the cluster")
	}

	// take the first one, we assume that all nodes are created in the same VPC
	name = getIDFromProvider(nodes.Items[0].Spec.ProviderID)

	params := &ec2.DescribeInstancesInput{
		Filters: []ec2.Filter{
			{
				Name: aws.String("instance-id"),
				Values: []string{
					name,
				},
			},
		},
	}
	log.Println("Trying to describe instance")
	req := svc.DescribeInstancesRequest(params)
	nodeInfo, err := req.Send(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "unable to describe AWS instance")
	}
	if len(nodeInfo.Reservations) == 0 {
		log.Println(err)
		return nil, fmt.Errorf("unable to describe AWS instance")
	}

	return nodeInfo, nil
}

// getSubnets returns a list of subnets within the VPC from the Kubernetes Node.
func getSubnets(nodeInfo *ec2.DescribeInstancesResponse, svc *ec2.Client, public bool) ([]string, error) {
	var result []string
	firstInstance := nodeInfo.Reservations[0].Instances[0]
	log.Printf("Taking subnets from node %v", *firstInstance.InstanceId)
	vpcID := firstInstance.VpcId
	for _, v := range firstInstance.SecurityGroups {
		log.Println("Security groupid: ", *v.GroupId)
	}
	log.Printf("Found VPC %v will search for subnet in that VPC\n", *vpcID)

	res := svc.DescribeSubnetsRequest(&ec2.DescribeSubnetsInput{Filters: []ec2.Filter{{Name: aws.String("vpc-id"), Values: []string{*vpcID}}}})
	subnets, err := res.Send(context.Background())

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to describe subnet in VPC %v", *vpcID))
	}
	for _, sn := range subnets.Subnets {
		if *sn.MapPublicIpOnLaunch == public {
			result = append(result, *sn.SubnetId)
		} else {
			log.Printf("Skipping subnet %v since it's public state was %v and we were looking for %v\n", *sn.SubnetId, *sn.MapPublicIpOnLaunch, public)
		}
	}

	log.Printf("Found the following subnets: ")
	for _, v := range result {
		log.Printf(v + " ")
	}
	return result, nil
}

func getIDFromProvider(x string) string {
	pos := strings.LastIndex(x, "/") + 1
	name := x[pos:]
	return name
}
func getSGS(kubectl *kubernetes.Clientset, svc *ec2.Client) ([]string, error) {

	nodes, err := kubectl.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get nodes")
	}
	name := ""

	if len(nodes.Items) > 0 {
		// take the first one, we assume that all nodes are created in the same VPC
		name = getIDFromProvider(nodes.Items[0].Spec.ProviderID)
	} else {
		return nil, fmt.Errorf("unable to find any nodes in the cluster")
	}
	log.Printf("Taking security groups from node %v", name)

	params := &ec2.DescribeInstancesInput{
		Filters: []ec2.Filter{
			{
				Name: aws.String("instance-id"),
				Values: []string{
					name,
				},
			},
		},
	}
	log.Println("Trying to describe instance")
	req := svc.DescribeInstancesRequest(params)
	res, err := req.Send(context.Background())
	if err != nil {
		log.Println(err)
		return nil, errors.Wrap(err, "Unable to describe AWS instance")
	}
	log.Println("Got instance response")

	var result []string
	if len(res.Reservations) >= 1 {
		for _, v := range res.Reservations[0].Instances[0].SecurityGroups {
			log.Println("Security Group Id: ", *v.GroupId)
			result = append(result, *v.GroupId)
		}
	}

	log.Printf("Found the follwing security groups: ")
	for _, v := range result {
		log.Printf(v + " ")
	}
	return result, nil
}

func ec2client(kubectl *kubernetes.Clientset) (*ec2.Client, error) {

	nodes, err := kubectl.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get nodes")
	}
	name := ""
	region := ""

	if len(nodes.Items) > 0 {
		// take the first one, we assume that all nodes are created in the same VPC
		name = getIDFromProvider(nodes.Items[0].Spec.ProviderID)
		region = nodes.Items[0].Labels["failure-domain.beta.kubernetes.io/region"]
	} else {
		return nil, fmt.Errorf("unable to find any nodes in the cluster")
	}
	log.Printf("Found node with ID: %v in region %v", name, region)

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	// Set the AWS Region that the service clients should use
	cfg.Region = region
	return ec2.New(cfg), nil
}
