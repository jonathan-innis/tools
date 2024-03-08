package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/samber/lo"
)

func main() {
	ctx := context.Background()
	cfg := lo.Must(config.LoadDefaultConfig(ctx))

	file := lo.Must(os.OpenFile("ami-ids.csv", os.O_RDWR|os.O_CREATE, 0777))
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	regions := []string{"us-west-2", "us-east-1", "us-east-2", "eu-west-1"}
	ec2Client := ec2.NewFromConfig(cfg)
	for _, version := range []string{"1.23", "1.24", "1.25", "1.26", "1.27", "1.28"} {
		rows := [][]string{}
		imageMap := map[string]map[string]string{}
		for _, region := range regions {
			out := lo.Must(ec2Client.DescribeImages(ctx, &ec2.DescribeImagesInput{
				Filters: []types.Filter{
					{
						Name: lo.ToPtr("name"),
						Values: []string{
							fmt.Sprintf("ubuntu-eks/k8s_%s/images/hvm-ssd/ubuntu-focal-20.04-arm64-server*", version),
							fmt.Sprintf("ubuntu-eks/k8s_%s/images/hvm-ssd/ubuntu-focal-20.04-amd64-server*", version),
						},
					},
				},
			}, WithRegion(region)))
			for _, img := range out.Images {
				if _, ok := imageMap[lo.FromPtr(img.Name)]; !ok {
					imageMap[lo.FromPtr(img.Name)] = map[string]string{}
				}
				imageMap[lo.FromPtr(img.Name)][region] = lo.FromPtr(img.ImageId)
			}
		}
		for k, v := range imageMap {
			row := []string{k}
			for _, r := range regions {
				row = append(row, v[r])
			}
			rows = append(rows, row)
		}
		sort.Slice(rows, func(i, j int) bool {
			return rows[i][0] < rows[j][0]
		})
		rows = append([][]string{{version}, append([]string{""}, regions...)}, rows...)
		rows = append(rows, []string{})
		w.WriteAll(rows)
	}
}

func WithRegion(region string) func(*ec2.Options) {
	return func(o *ec2.Options) {
		o.Region = region
	}
}
