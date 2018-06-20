package main

import (
	"database/sql"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticache"
	_ "github.com/lib/pq"
	"github.com/nu7hatch/gouuid"
	"log"
	"os"
	"strconv"
	"strings"
)

func provision(plan string) (error, string) {
	small_parameter_group := os.Getenv("SMALL_PARAMETER_GROUP")
	medium_parameter_group := os.Getenv("MEDIUM_PARAMETER_GROUP")
	large_parameter_group := os.Getenv("LARGE_PARAMETER_GROUP")
	aws_region := os.Getenv("AWS_REGION")
	engine_version := os.Getenv("ENGINE_VERSION")
	subnet_group := os.Getenv("SUBNET_GROUP")

	if small_parameter_group == "" {
		small_parameter_group = "redis-32-small"
	}
	if medium_parameter_group == "" {
		medium_parameter_group = "redis-32-medium"
	}
	if large_parameter_group == "" {
		large_parameter_group = "redis-32-large"
	}
	if aws_region == "" {
		aws_region = "us-west-2"
	}
	if engine_version == "" {
		engine_version = "3.2.10"
	}
	if subnet_group == "" {
		subnet_group = "redis-subnet-group"
	}

	cacheparametergroupname := small_parameter_group
	cachenodetype := os.Getenv("SMALL_INSTANCE_TYPE")
	numcachenodes := int64(1)
	billingcode := "pre-provisioned"
	u, err := uuid.NewV4()
	if err != nil {
		return err, ""
	}
	name := os.Getenv("NAME_PREFIX") + "-" + strings.Split(u.String(), "-")[0]

	if plan == "small" {
		cacheparametergroupname = small_parameter_group
		cachenodetype = os.Getenv("SMALL_INSTANCE_TYPE")
		numcachenodes = int64(1)
	} else if plan == "medium" {
		cacheparametergroupname = medium_parameter_group
		cachenodetype = os.Getenv("MEDIUM_INSTANCE_TYPE")
		numcachenodes = int64(1)
	} else if plan == "large" {
		cacheparametergroupname = large_parameter_group
		cachenodetype = os.Getenv("LARGE_INSTANCE_TYPE")
		numcachenodes = int64(1)
	} else {
		return errors.New("Invalid plan specified: " + plan), ""
	}

	svc := elasticache.New(session.New(&aws.Config{
		Region: aws.String(aws_region),
	}))

	params := &elasticache.CreateCacheClusterInput{
		CacheClusterId:          aws.String(name),
		AutoMinorVersionUpgrade: aws.Bool(true),
		CacheNodeType:           aws.String(cachenodetype),
		CacheParameterGroupName: aws.String(cacheparametergroupname),
		CacheSubnetGroupName:    aws.String(subnet_group),
		Engine:                  aws.String("redis"),
		EngineVersion:           aws.String(engine_version),
		NumCacheNodes:           aws.Int64(numcachenodes),
		Port:                    aws.Int64(6379),
		SecurityGroupIds: []*string{
			aws.String(os.Getenv("ELASTICACHE_SECURITY_GROUP")),
		},
		Tags: []*elasticache.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(name),
			},
			{
				Key:   aws.String("billingcode"),
				Value: aws.String(billingcode),
			},
		},
	}
	_, err = svc.CreateCacheCluster(params)
	if err != nil {
		return err, ""
	}
	return nil, name
}

func insertnew(name string, plan string, claimed string) {
	uri := os.Getenv("BROKER_DB")
	db, err := sql.Open("postgres", uri)
	if err != nil {
		log.Fatalln(err)
	}
	var newname string
	err = db.QueryRow("INSERT INTO provision(name,plan,claimed) VALUES($1,$2,$3) returning name;", name, plan, claimed).Scan(&newname)
	if err != nil {
		log.Fatalln(err)
	}
	err = db.Close()
}

func main() {
	uri := os.Getenv("BROKER_DB")
	db, err := sql.Open("postgres", uri)
	if err != nil {
		log.Fatalln(err)
	}

	newname := "new"
	provisionsmall, _ := strconv.Atoi(os.Getenv("PROVISION_SMALL"))
	provisionmedium, _ := strconv.Atoi(os.Getenv("PROVISION_MEDIUM"))
	provisionlarge, _ := strconv.Atoi(os.Getenv("PROVISION_LARGE"))
	var smallcount int
	err = db.QueryRow("SELECT count(*) as smallcount from provision where plan='small' and claimed='no'").Scan(&smallcount)
	if err != nil {
		log.Fatalln(err)
	}

	if smallcount < provisionsmall {
		err, newname = provision("small")
		if err != nil {
			log.Fatalln(err)
		}
		insertnew(newname, "small", "no")
	}

	var mediumcount int
	err = db.QueryRow("SELECT count(*) as mediumcount from provision where plan='medium' and claimed='no'").Scan(&mediumcount)
	if err != nil {
		log.Fatalln(err)
	}

	if mediumcount < provisionmedium {
		err, newname = provision("medium")
		if err != nil {
			log.Fatalln(err)
		}
		insertnew(newname, "medium", "no")
	}

	var largecount int
	err = db.QueryRow("SELECT count(*) as largecount from provision where plan='large' and claimed='no'").Scan(&largecount)
	if err != nil {
		log.Fatalln(err)
	}

	if largecount < provisionlarge {
		err, newname = provision("large")
		if err != nil {
			log.Fatalln(err)
		}
		insertnew(newname, "large", "no")
	}
	err = db.Close()
}
