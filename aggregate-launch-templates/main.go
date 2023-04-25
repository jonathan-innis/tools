package main

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/samber/lo"
)

var (
	groupBy string

	groupByOptions = []string{"name", "type"}
)

func init() {
	flag.StringVar(&groupBy, "group-by", "type", "--group-by is the way to group the launch templates")
	flag.Parse()
}

func main() {
	if !lo.Contains(groupByOptions, groupBy) {
		panic(fmt.Sprintf("--group-by option %q is not a valid option", groupBy))
	}
	sess := session.Must(session.NewSession(
		request.WithRetryer(
			&aws.Config{STSRegionalEndpoint: endpoints.RegionalSTSEndpoint},
			client.DefaultRetryer{NumMaxRetries: client.DefaultRetryerMaxNumRetries},
		),
	))
	ec2api := ec2.New(sess)
	var launchTemplates []*ec2.LaunchTemplate
	lo.Must0(ec2api.DescribeLaunchTemplatesPages(&ec2.DescribeLaunchTemplatesInput{}, func(out *ec2.DescribeLaunchTemplatesOutput, _ bool) bool {
		launchTemplates = append(launchTemplates, out.LaunchTemplates...)
		return true
	}))

	var grouped map[string][]*ec2.LaunchTemplate
	switch groupBy {
	case "name":
		grouped = groupByName(launchTemplates)
	default:
		grouped = groupByType(launchTemplates)
	}

	var res []lo.Tuple2[string, int]
	for k, v := range grouped {
		res = append(res, lo.Tuple2[string, int]{A: k, B: len(v)})
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].B > res[j].B
	})
	fmt.Printf("All Launch Templates: %d\n", len(launchTemplates))
	for _, elem := range res {
		fmt.Printf("%s: %d\n", elem.A, elem.B)
	}
}

func groupByName(launchTemplates []*ec2.LaunchTemplate) map[string][]*ec2.LaunchTemplate {
	return lo.GroupBy(launchTemplates, func(lt *ec2.LaunchTemplate) string {
		if t, ok := lo.Find(lt.Tags, func(t *ec2.Tag) bool {
			return aws.StringValue(t.Key) == "karpenter.k8s.aws/cluster"
		}); ok {
			return aws.StringValue(t.Value)
		}
		return ""
	})
}

func groupByType(launchTemplates []*ec2.LaunchTemplate) map[string][]*ec2.LaunchTemplate {
	knownTestSuites := []string{"integration", "upgrade", "machine", "chaos", "drift", "utilization", "consolidation", "ipv6", "interruption"}
	return lo.GroupBy(launchTemplates, func(lt *ec2.LaunchTemplate) string {
		if t, ok := lo.Find(lt.Tags, func(t *ec2.Tag) bool {
			return aws.StringValue(t.Key) == "karpenter.k8s.aws/cluster"
		}); ok {
			for _, suite := range knownTestSuites {
				if strings.Contains(aws.StringValue(t.Value), suite) {
					return suite
				}
			}
			return aws.StringValue(t.Value)
		}
		return ""
	})
}
